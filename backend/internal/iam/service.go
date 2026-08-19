// Package iam 实现认证、会话与 RBAC 决策（Phase 2）。
package iam

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dxcloud/cloud-api/internal/config"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/pkg/jwt"
	"github.com/dxcloud/cloud-api/pkg/redact"
	"github.com/dxcloud/cloud-api/pkg/redisx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	refreshKeyPrefix = "auth:refresh:"
	denyKeyPrefix    = "auth:deny:"
	permsCachePrefix = "cache:perms:"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrRefreshInvalid     = errors.New("invalid refresh token")
	ErrAccountLocked      = errors.New("account locked: too many failed attempts")
)

type Service struct {
	cfg *config.Config
	log *zap.Logger
	db  *gorm.DB
	rdb *redis.Client
	r   *repository.Repos
}

func NewService(cfg *config.Config, log *zap.Logger, db *gorm.DB, rdb *redis.Client, r *repository.Repos) *Service {
	return &Service{cfg: cfg, log: log, db: db, rdb: rdb, r: r}
}

// ---------- 密码 ----------

func hashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func verifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// ---------- 令牌 ----------

func (s *Service) accessTTL() time.Duration {
	d, err := time.ParseDuration(s.cfg.JWT.AccessTTL)
	if err != nil || d <= 0 {
		d = 15 * time.Minute
	}
	return d
}

func (s *Service) refreshTTL() time.Duration {
	d, err := time.ParseDuration(s.cfg.JWT.RefreshTTL)
	if err != nil || d <= 0 {
		d = 7 * 24 * time.Hour
	}
	return d
}

func newRefreshValue() string {
	return uuid.NewString() + uuid.NewString()
}

// IssueTokens 为指定用户签发 Access + Refresh。
func (s *Service) IssueTokens(ctx context.Context, user *model.User) (access, refresh string, expiresIn int64, err error) {
	access, jti, err := jwt.Generate(s.cfg.JWT.Secret, s.accessTTL(), user.ID, user.Username)
	if err != nil {
		return "", "", 0, err
	}
	_ = jti
	refresh = newRefreshValue()
	fields := map[string]any{
		"user_id":    fmt.Sprintf("%d", user.ID),
		"username":   user.Username,
		"created_at": time.Now().Format(time.RFC3339),
	}
	if err := redisx.HSet(ctx, s.rdb, refreshKeyPrefix+refresh, fields); err != nil {
		return "", "", 0, err
	}
	if err := redisx.Expire(ctx, s.rdb, refreshKeyPrefix+refresh, s.refreshTTL()); err != nil {
		return "", "", 0, err
	}
	return access, refresh, int64(s.accessTTL().Seconds()), nil
}

// Authenticate 校验 Access Token（签名 → 黑名单 → 用户状态）。
func (s *Service) Authenticate(ctx context.Context, tokenStr string) (*model.User, *jwt.Claims, error) {
	claims, err := jwt.Parse(s.cfg.JWT.Secret, tokenStr)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	if claims.ID == "" {
		return nil, nil, ErrInvalidCredentials
	}
	denied, err := redisx.Exists(ctx, s.rdb, denyKeyPrefix+claims.ID)
	if err != nil {
		return nil, nil, err
	}
	if denied {
		return nil, nil, ErrInvalidCredentials
	}
	user, err := s.r.UserGetByID(claims.UserID)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	if user.Status != model.UserStatusActive {
		return nil, nil, ErrUserDisabled
	}
	return user, claims, nil
}

// Refresh 旋转刷新令牌：旧 token 一次性作废，签发新对。
func (s *Service) Refresh(ctx context.Context, refreshToken string) (access, newRefresh string, expiresIn int64, err error) {
	key := refreshKeyPrefix + refreshToken
	m, err := redisx.HGetAll(ctx, s.rdb, key)
	if err != nil || len(m) == 0 {
		return "", "", 0, ErrRefreshInvalid
	}
	var userID uint64
	if _, err := fmt.Sscanf(m["user_id"], "%d", &userID); err != nil {
		return "", "", 0, ErrRefreshInvalid
	}
	user, err := s.r.UserGetByID(userID)
	if err != nil || user.Status != model.UserStatusActive {
		return "", "", 0, ErrRefreshInvalid
	}
	// 旧 refresh 立即作废（旋转）
	_ = redisx.Del(ctx, s.rdb, key)
	access, newRefresh, expiresIn, err = s.IssueTokens(ctx, user)
	if err != nil {
		return "", "", 0, err
	}
	return access, newRefresh, expiresIn, nil
}

// Logout 黑名单当前 Access（至其过期）并删除 Refresh。
func (s *Service) Logout(ctx context.Context, accessJTI, refreshToken string) {
	if accessJTI != "" {
		claimsRemain := s.accessTTL()
		_ = redisx.Set(ctx, s.rdb, denyKeyPrefix+accessJTI, "1", claimsRemain)
	}
	if refreshToken != "" {
		_ = redisx.Del(ctx, s.rdb, refreshKeyPrefix+refreshToken)
	}
}

// RevokeAllSessions 撤销用户全部 Refresh 会话（改密/禁用时调用）。
func (s *Service) RevokeAllSessions(ctx context.Context, userID uint64) {
	keys, err := redisx.ScanKeys(ctx, s.rdb, refreshKeyPrefix+"*")
	if err != nil {
		s.log.Warn("scan refresh keys failed", zap.Error(err))
		return
	}
	for _, k := range keys {
		uid, _ := redisx.HGet(ctx, s.rdb, k, "user_id")
		if uid == fmt.Sprintf("%d", userID) {
			_ = redisx.Del(ctx, s.rdb, k)
		}
	}
}

type SessionInfo struct {
	JTI       string    `json:"jti"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresIn int64     `json:"expires_in"`
}

// Sessions 列出当前用户的活跃会话。
func (s *Service) Sessions(ctx context.Context, userID uint64) ([]SessionInfo, error) {
	keys, err := redisx.ScanKeys(ctx, s.rdb, refreshKeyPrefix+"*")
	if err != nil {
		return nil, err
	}
	uidStr := fmt.Sprintf("%d", userID)
	var out []SessionInfo
	for _, k := range keys {
		uid, _ := redisx.HGet(ctx, s.rdb, k, "user_id")
		if uid != uidStr {
			continue
		}
		created, _ := redisx.HGet(ctx, s.rdb, k, "created_at")
		ttl, _ := redisx.TTL(ctx, s.rdb, k)
		info := SessionInfo{JTI: strings.TrimPrefix(k, refreshKeyPrefix), ExpiresIn: int64(ttl.Seconds())}
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			info.CreatedAt = t
		}
		out = append(out, info)
	}
	return out, nil
}

// DeleteSession 撤销指定会话（校验归属）。
func (s *Service) DeleteSession(ctx context.Context, userID uint64, jti string) error {
	key := refreshKeyPrefix + jti
	uid, err := redisx.HGet(ctx, s.rdb, key, "user_id")
	if err != nil {
		return ErrRefreshInvalid
	}
	if uid != fmt.Sprintf("%d", userID) {
		return ErrRefreshInvalid
	}
	return redisx.Del(ctx, s.rdb, key)
}

// ---------- 认证业务 ----------

func (s *Service) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	if len(username) < 3 || len(username) > 32 {
		return nil, errors.New("用户名长度需为 3-32 位")
	}
	if len(password) < 8 || len(password) > 72 {
		return nil, errors.New("密码长度需为 8-72 位")
	}
	if _, err := s.r.UserGetByUsername(username); err == nil {
		return nil, errors.New("用户名已存在")
	}
	if _, err := s.r.UserGetByEmail(email); err == nil {
		return nil, errors.New("邮箱已被注册")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{Username: username, Email: email, PasswordHash: hash, Nickname: username, Status: model.UserStatusActive}
	if err := s.r.UserCreate(user); err != nil {
		return nil, err
	}
	// 默认角色：user
	if role, err := s.r.RoleGetByCode("user"); err == nil {
		_ = s.r.UserRoleReplace(user.ID, []uint64{role.ID})
	}
	return user, nil
}

const (
	loginFailKeyPrefix = "dx:login:fail:"
	loginMaxFail       = 5
	loginLockTTL       = 15 * time.Minute
)

func (s *Service) Login(ctx context.Context, username, password, ip, ua string) (*model.User, error) {
	// 防爆破：连续失败 5 次锁定 15 分钟（Redis 计数，键按用户名小写归一）
	failKey := loginFailKeyPrefix + strings.ToLower(username)
	if v, err := redisx.Get(ctx, s.rdb, failKey); err == nil {
		if n, err2 := strconv.Atoi(v); err2 == nil && n >= loginMaxFail {
			s.log.Warn("login blocked: account locked", zap.String("username", username), zap.String("ip", ip))
			return nil, ErrAccountLocked
		}
	}
	user, err := s.r.UserGetByUsername(username)
	fail := func(msg string, status int8, uid *uint64) (*model.User, error) {
		n, _ := redisx.Incr(ctx, s.rdb, failKey)
		if n == 1 {
			_ = redisx.Expire(ctx, s.rdb, failKey, loginLockTTL)
		}
		s.log.Warn("login failed", zap.String("username", username), zap.Int64("fail_count", n), zap.String("msg", msg))
		_ = s.r.LoginLogCreate(&model.LoginLog{UserID: uid, IP: ip, UserAgent: ua, Status: status, Message: msg})
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return fail("user not found", 2, nil)
	}
	if user.Status != model.UserStatusActive {
		return fail("user disabled", 2, &user.ID)
	}
	if !verifyPassword(user.PasswordHash, password) {
		return fail("wrong password", 2, &user.ID)
	}
	_ = redisx.Del(ctx, s.rdb, failKey)
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ip
	_ = s.r.DB.Model(user).Select("last_login_at", "last_login_ip").Updates(map[string]any{
		"last_login_at": now, "last_login_ip": ip,
	}).Error
	_ = s.r.LoginLogCreate(&model.LoginLog{UserID: &user.ID, IP: ip, UserAgent: ua, Status: 1, Message: "success"})
	return user, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uint64, oldPwd, newPwd string) error {
	if len(newPwd) < 8 || len(newPwd) > 72 {
		return errors.New("新密码长度需为 8-72 位")
	}
	user, err := s.r.UserGetByID(userID)
	if err != nil {
		return err
	}
	if !verifyPassword(user.PasswordHash, oldPwd) {
		return errors.New("原密码不正确")
	}
	hash, err := hashPassword(newPwd)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	if err := s.r.UserUpdate(user); err != nil {
		return err
	}
	// 改密后撤销其他全部会话
	s.RevokeAllSessions(ctx, userID)
	return nil
}

// UpdateProfile 更新当前用户昵称与头像（仅本人可改）。
func (s *Service) UpdateProfile(ctx context.Context, userID uint64, nickname, avatarURL string) error {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return errors.New("昵称不能为空")
	}
	if utf8.RuneCountInString(nickname) > 32 {
		return errors.New("昵称最长 32 个字符")
	}
	if len(avatarURL) > 32768 {
		return errors.New("头像数据过大（≤32KB）")
	}
	user, err := s.r.UserGetByID(userID)
	if err != nil {
		return err
	}
	user.Nickname = nickname
	user.AvatarURL = avatarURL
	return s.r.UserUpdate(user)
}

// ---------- RBAC ----------

func (s *Service) GetUserRoleCodes(ctx context.Context, userID uint64) ([]string, error) {
	return s.r.UserRoleCodes(userID)
}

func (s *Service) IsSuperAdmin(ctx context.Context, userID uint64) (bool, error) {
	codes, err := s.GetUserRoleCodes(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, c := range codes {
		if c == "superadmin" {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermCodes 返回用户有效权限码（Redis 缓存 5 分钟；superadmin 返回全部）。
func (s *Service) GetUserPermCodes(ctx context.Context, userID uint64) ([]string, error) {
	cacheKey := permsCachePrefix + fmt.Sprintf("%d", userID)
	if v, err := redisx.Get(ctx, s.rdb, cacheKey); err == nil && v != "" {
		return strings.Split(v, ","), nil
	}
	isSA, err := s.IsSuperAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	var codes []string
	if isSA {
		perms, err := s.r.PermissionList()
		if err != nil {
			return nil, err
		}
		for _, p := range perms {
			codes = append(codes, p.Code)
		}
	} else {
		roleIDs, err := s.r.UserRoleIDs(userID)
		if err != nil {
			return nil, err
		}
		codes, err = s.r.RolePermCodes(roleIDs)
		if err != nil {
			return nil, err
		}
	}
	if len(codes) > 0 {
		_ = redisx.Set(ctx, s.rdb, cacheKey, strings.Join(codes, ","), 5*time.Minute)
	}
	return codes, nil
}

// InvalidatePermCache 角色/权限变更后调用。
func (s *Service) InvalidatePermCache(ctx context.Context, userID uint64) {
	_ = redisx.Del(ctx, s.rdb, permsCachePrefix+fmt.Sprintf("%d", userID))
}

// HasPerm 判断用户是否拥有某权限。
func (s *Service) HasPerm(ctx context.Context, userID uint64, perm string) (bool, error) {
	codes, err := s.GetUserPermCodes(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, c := range codes {
		if c == perm {
			return true, nil
		}
	}
	return false, nil
}

// ---------- 审计 ----------

// Audit 写入审计日志（status: 1=成功 2=拒绝）。
func (s *Service) Audit(ctx context.Context, userID *uint64, action, resourceType, resourceID, ip, requestID string, status int8, detail any) {
	detailStr := ""
	if detail != nil {
		// 敏感字段脱敏（Phase 11）：password/token/secret 等键值一律 ***
		if mm, ok := detail.(map[string]any); ok {
			detail = redact.Map(mm)
		}
		if b, err := jsonMarshal(detail); err == nil {
			detailStr = b
		}
	}
	log := &model.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IP:           ip,
		RequestID:    requestID,
		Detail:       detailStr,
		Status:       status,
	}
	if err := s.r.AuditLogCreate(log); err != nil {
		s.log.Warn("audit log write failed", zap.Error(err), zap.String("action", action))
	}
}
