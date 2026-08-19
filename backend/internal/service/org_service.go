// Package service：组织 / 配额 / 计费服务（Phase 10）。
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"go.uber.org/zap"
)

// 默认配额（免费套餐）
var defaultQuotas = map[string]int64{
	"ecs_count": 5,
	"cpu":       8,
	"memory":    16 * 1024,
	"storage":   100,
	"network":   10,
	"pipeline":  10,
}

type OrgService struct {
	repo   *repository.Repos
	iamSvc *iam.Service
}

func NewOrgService(repo *repository.Repos, iamSvc *iam.Service) *OrgService {
	return &OrgService{repo: repo, iamSvc: iamSvc}
}

// Create 创建组织：创建者自动成为 owner 成员，并写入默认配额 + 初始虚拟余额。
func (s *OrgService) Create(ctx context.Context, ac AccessCtx, name, code, plan, ip, requestID string) (*model.Organization, error) {
	if name == "" || len(name) > 127 {
		return nil, errors.New("组织名需 1-127 位")
	}
	if code == "" {
		code = name
	}
	if plan == "" {
		plan = "free"
	}
	org := &model.Organization{Name: name, Code: code, Plan: plan, Credit: 1000, Status: 1, CreatedBy: &ac.UserID}
	if err := s.repo.OrgCreate(org); err != nil {
		return nil, err
	}
	if err := s.repo.OrgMemberCreate(&model.OrganizationMember{
		OrgID: org.ID, UserID: ac.UserID, OrgRole: "owner", Status: 1,
	}); err != nil {
		return nil, err
	}
	for rt, limit := range defaultQuotas {
		_ = s.repo.QuotaUpsert(&model.ResourceQuota{OrgID: org.ID, ResourceType: rt, LimitValue: limit})
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "org.create", "org", fmt.Sprintf("%d", org.ID), ip, requestID, 1, map[string]any{"name": name})
	return org, nil
}

func (s *OrgService) Delete(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	org, err := s.repo.OrgGetByID(id)
	if err != nil {
		return err
	}
	if !ac.hasRole("superadmin") {
		role, ok := s.IsMember(ctx, id, ac.UserID)
		if !ok || (role != "owner" && role != "admin") {
			return ErrForbidden
		}
	}
	if err := s.repo.OrgSoftDelete(id); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "org.delete", "org", org.Name, ip, requestID, 1, nil)
	return nil
}

// IsMember 判断用户是否组织成员（返回成员角色）。
func (s *OrgService) IsMember(ctx context.Context, orgID, userID uint64) (string, bool) {
	if orgID == 0 {
		return "", true // 未指定组织（单租户兼容模式）
	}
	m, err := s.repo.OrgMemberGet(orgID, userID)
	if err != nil {
		return "", false
	}
	return m.OrgRole, m.Status == 1
}

func (s *OrgService) ListOrgs() ([]model.Organization, error) {
	return s.repo.OrgList()
}

// MemberOrgs 用户所属组织（供前端组织切换器）。
func (s *OrgService) MemberOrgs(userID uint64) ([]model.Organization, error) {
	return s.repo.OrgListByMember(userID)
}

// ---------- 成员管理 ----------

func (s *OrgService) Members(ctx context.Context, ac AccessCtx, orgID uint64, ip, requestID string) ([]model.OrganizationMember, error) {
	if role, ok := s.IsMember(ctx, orgID, ac.UserID); !ok && !ac.hasRole("superadmin") {
		_ = role
		return nil, ErrForbidden
	}
	return s.repo.OrgMemberList(orgID)
}

func (s *OrgService) AddMember(ctx context.Context, ac AccessCtx, orgID uint64, username, orgRole, ip, requestID string) error {
	role, ok := s.IsMember(ctx, orgID, ac.UserID)
	if !ok && !ac.hasRole("superadmin") {
		return ErrForbidden
	}
	if !ac.hasRole("superadmin") && role != "owner" && role != "admin" {
		return ErrForbidden
	}
	if orgRole == "" || (orgRole != "owner" && orgRole != "admin" && orgRole != "member" && orgRole != "viewer") {
		return errors.New("org_role 需为 owner/admin/member/viewer")
	}
	user, err := s.repo.UserGetByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}
	if _, err := s.repo.OrgMemberGet(orgID, user.ID); err == nil {
		return errors.New("user already in org")
	}
	if err := s.repo.OrgMemberCreate(&model.OrganizationMember{
		OrgID: orgID, UserID: user.ID, OrgRole: orgRole, Status: 1,
	}); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "org.member.add", "org", fmt.Sprintf("%d", orgID), ip, requestID, 1, map[string]any{"user": username, "role": orgRole})
	return nil
}

func (s *OrgService) RemoveMember(ctx context.Context, ac AccessCtx, orgID, memberID uint64, ip, requestID string) error {
	role, ok := s.IsMember(ctx, orgID, ac.UserID)
	if !ok && !ac.hasRole("superadmin") {
		return ErrForbidden
	}
	if !ac.hasRole("superadmin") && role != "owner" {
		return ErrForbidden
	}
	if memberID == ac.UserID {
		return errors.New("cannot remove yourself")
	}
	if err := s.repo.OrgMemberDelete(orgID, memberID); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "org.member.remove", "org", fmt.Sprintf("%d", orgID), ip, requestID, 1, nil)
	return nil
}

// ---------- 配额 ----------

type QuotaService struct {
	repo *repository.Repos
}

func NewQuotaService(repo *repository.Repos) *QuotaService {
	return &QuotaService{repo: repo}
}

func (s *QuotaService) List(orgID uint64) ([]model.ResourceQuota, error) {
	if orgID == 0 {
		// 单租户模式：返回默认配额视图
		out := make([]model.ResourceQuota, 0, len(defaultQuotas))
		for rt, limit := range defaultQuotas {
			out = append(out, model.ResourceQuota{ResourceType: rt, LimitValue: limit})
		}
		return out, nil
	}
	return s.repo.QuotaList(orgID)
}

func (s *QuotaService) Update(orgID uint64, resourceType string, limit int64) error {
	if orgID == 0 {
		return errors.New("单租户模式使用内置默认配额，不支持修改")
	}
	if _, ok := defaultQuotas[resourceType]; !ok {
		return fmt.Errorf("未知配额类型 %s（ecs_count/cpu/memory/storage/network/pipeline）", resourceType)
	}
	if limit < 0 || limit > 100000 {
		return errors.New("配额值非法")
	}
	return s.repo.QuotaUpsert(&model.ResourceQuota{OrgID: orgID, ResourceType: resourceType, LimitValue: limit})
}

// CheckEcsQuota ECS 创建配额校验（组织配额优先，无组织回退单租户默认）。
func (s *QuotaService) CheckEcsQuota(orgID, ownerID uint64, addCPU float64, addMem int64) error {
	maxInstances, maxCPU, maxMem := int64(5), int64(8), int64(16*1024)
	if orgID > 0 {
		if q, err := s.repo.QuotaGet(orgID, "ecs_count"); err == nil && q.LimitValue > 0 {
			maxInstances = q.LimitValue
		}
		if q, err := s.repo.QuotaGet(orgID, "cpu"); err == nil && q.LimitValue > 0 {
			maxCPU = q.LimitValue
		}
		if q, err := s.repo.QuotaGet(orgID, "memory"); err == nil && q.LimitValue > 0 {
			maxMem = q.LimitValue
		}
	}
	var count int64
	var cpu float64
	var mem int64
	var err error
	if orgID > 0 {
		count, cpu, mem, err = s.repo.EcsOrgQuotaUsage(orgID)
	} else {
		count, cpu, mem, err = s.repo.EcsQuotaUsage(ownerID)
	}
	if err != nil {
		return err
	}
	if count+1 > maxInstances {
		return fmt.Errorf("%w：实例数已达上限（%d/%d 个），可清理闲置实例或在「组织与配额」页提升配额", ErrQuotaExceed, count, maxInstances)
	}
	if cpu+addCPU > float64(maxCPU) {
		return fmt.Errorf("%w：CPU 已达上限（需 %.2f 核，剩余 %.2f/%d 核），可降低规格或提升配额", ErrQuotaExceed, cpu+addCPU, float64(maxCPU)-cpu, maxCPU)
	}
	if mem+addMem > maxMem {
		return fmt.Errorf("%w：内存已达上限（需 %d MB，剩余 %d/%d MB），可降低规格或提升配额", ErrQuotaExceed, mem+addMem, maxMem-mem, maxMem)
	}
	return nil
}

// ---------- 计费 ----------

type BillingService struct {
	repo *repository.Repos
	log  *zap.Logger
}

func NewBillingService(repo *repository.Repos, log *zap.Logger) *BillingService {
	return &BillingService{repo: repo, log: log}
}

// Collect 每小时用量结算（虚拟计费）：对运行中实例按其规格折算 cpu_hour/mem_gb_hour/disk_gb_hour，
// 并从组织余额扣费。价格（虚拟）：CPU ¥0.1/核时、内存 ¥0.05/GB时、磁盘 ¥0.01/GB时。
func (s *BillingService) Collect(ctx context.Context) error {
	period := time.Now().Truncate(time.Hour)
	instances, _, err := s.repo.EcsList(repository.EcsFilter{Status: "running", Page: 1, Size: 1000})
	if err != nil {
		return err
	}
	type orgUsage struct {
		cpuH, memH, diskH float64
	}
	byOrg := map[uint64]*orgUsage{}
	for i := range instances {
		inst := &instances[i]
		orgID := uint64(0)
		if inst.OrgID != nil {
			orgID = *inst.OrgID
		}
		u := byOrg[orgID]
		if u == nil {
			u = &orgUsage{}
			byOrg[orgID] = u
		}
		u.cpuH += inst.CPU
		u.memH += float64(inst.MemoryMB) / 1024
		u.diskH += float64(inst.DiskGB)
	}
	for orgID, u := range byOrg {
		rows := []model.ResourceUsage{
			{OrgID: orgID, ResourceType: "cpu_hour", UsedValue: round2(u.cpuH), Period: period},
			{OrgID: orgID, ResourceType: "mem_gb_hour", UsedValue: round2(u.memH), Period: period},
			{OrgID: orgID, ResourceType: "disk_gb_hour", UsedValue: round2(u.diskH), Period: period},
		}
		for i := range rows {
			if rows[i].UsedValue <= 0 {
				continue
			}
			if err := s.repo.UsageCreate(&rows[i]); err != nil {
				s.log.Warn("usage create failed", zap.Error(err))
			}
		}
		cost := u.cpuH*0.1 + u.memH*0.05 + u.diskH*0.01
		if orgID > 0 {
			s.debit(ctx, orgID, cost)
		}
	}
	return nil
}

func (s *BillingService) debit(ctx context.Context, orgID uint64, cost float64) {
	// 原子扣费：UPDATE organizations SET credit = credit - ? WHERE id = ?
	if err := s.repo.OrgAdjustCredit(orgID, -cost); err != nil {
		s.log.Warn("debit failed", zap.Uint64("org", orgID), zap.Float64("cost", cost), zap.Error(err))
	}
}

// OrgByID 组织查询。
func (s *BillingService) OrgByID(orgID uint64) (*model.Organization, error) {
	if orgID == 0 {
		return &model.Organization{Credit: 0}, nil
	}
	return s.repo.OrgGetByID(orgID)
}

// UsageList / UsageSum 透传。
func (s *BillingService) UsageList(orgID uint64, limit int) ([]model.ResourceUsage, error) {
	return s.repo.UsageList(orgID, time.Time{}, time.Time{}, limit)
}

func (s *BillingService) UsageSum(orgID uint64, from, to time.Time) (map[string]float64, error) {
	return s.repo.UsageSum(orgID, from, to)
}

// Recharge 虚拟充值（管理员）。
func (s *BillingService) Recharge(ctx context.Context, orgID uint64, amount float64) error {
	if amount <= 0 || amount > 1000000 {
		return errors.New("充值金额非法")
	}
	if _, err := s.repo.OrgGetByID(orgID); err != nil {
		return err
	}
	// 原子充值：UPDATE organizations SET credit = credit + ? WHERE id = ?
	return s.repo.OrgAdjustCredit(orgID, amount)
}

// HasCredit 余额检查（<=0 且用量>0 时拒绝新资源创建）。
func (s *BillingService) HasCredit(orgID uint64) bool {
	if orgID == 0 {
		return true
	}
	org, err := s.repo.OrgGetByID(orgID)
	if err != nil {
		return true
	}
	return org.Credit > -1000 // 允许适度透支，管理员可充值
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
