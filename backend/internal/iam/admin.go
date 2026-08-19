package iam

import (
	"context"
	"errors"
	"regexp"

	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
)

var roleCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,31}$`)

// Me 返回用户 + 角色码 + 权限码。
func (s *Service) Me(ctx context.Context, userID uint64) (*model.User, []string, []string, []string, error) {
	user, err := s.r.UserGetByID(userID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	roles, err := s.GetUserRoleCodes(ctx, userID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	roleNames, err := s.r.UserRoleNames(userID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	perms, err := s.GetUserPermCodes(ctx, userID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return user, roles, roleNames, perms, nil
}

func (s *Service) resolveRoleIDs(roleCodes []string, defCode string) ([]uint64, error) {
	if len(roleCodes) == 0 {
		roleCodes = []string{defCode}
	}
	var ids []uint64
	for _, code := range roleCodes {
		role, err := s.r.RoleGetByCode(code)
		if err != nil {
			return nil, errors.New("unknown role: " + code)
		}
		ids = append(ids, role.ID)
	}
	return ids, nil
}

// ---------- 用户管理（Admin） ----------

func (s *Service) CreateUser(ctx context.Context, username, email, password, nickname string, roleCodes []string) (*model.User, error) {
	if len(username) < 3 || len(username) > 32 {
		return nil, errors.New("username length must be 3-32")
	}
	if len(password) < 8 || len(password) > 72 {
		return nil, errors.New("password length must be 8-72")
	}
	if _, err := s.r.UserGetByUsername(username); err == nil {
		return nil, errors.New("username already exists")
	}
	if _, err := s.r.UserGetByEmail(email); err == nil {
		return nil, errors.New("email already exists")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	if nickname == "" {
		nickname = username
	}
	user := &model.User{
		Username: username, Email: email, PasswordHash: hash,
		Nickname: nickname, Status: model.UserStatusActive,
	}
	if err := s.r.UserCreate(user); err != nil {
		return nil, err
	}
	roleIDs, err := s.resolveRoleIDs(roleCodes, "user")
	if err != nil {
		return nil, err
	}
	if err := s.r.UserRoleReplace(user.ID, roleIDs); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uint64, nickname string, status *int8) error {
	user, err := s.r.UserGetByID(id)
	if err != nil {
		return err
	}
	if nickname != "" {
		user.Nickname = nickname
	}
	if status != nil {
		user.Status = *status
		if user.Status != model.UserStatusActive {
			s.RevokeAllSessions(ctx, id)
		}
	}
	return s.r.UserUpdate(user)
}

// DeleteUser 删除用户（保护：不能删自己、不能删 superadmin）。
func (s *Service) DeleteUser(ctx context.Context, operatorID, targetID uint64) error {
	if operatorID == targetID {
		return errors.New("cannot delete yourself")
	}
	roles, err := s.GetUserRoleCodes(ctx, targetID)
	if err != nil {
		return err
	}
	for _, rc := range roles {
		if rc == "superadmin" {
			return errors.New("cannot delete superadmin")
		}
	}
	if err := s.r.UserSoftDelete(targetID); err != nil {
		return err
	}
	s.RevokeAllSessions(ctx, targetID)
	s.InvalidatePermCache(ctx, targetID)
	return nil
}

// GrantRoles 替换用户角色集。
func (s *Service) GrantRoles(ctx context.Context, userID uint64, roleCodes []string) error {
	roleIDs, err := s.resolveRoleIDs(roleCodes, "user")
	if err != nil {
		return err
	}
	if err := s.r.UserRoleReplace(userID, roleIDs); err != nil {
		return err
	}
	s.InvalidatePermCache(ctx, userID)
	return nil
}

// ---------- 角色/权限管理（Admin） ----------

func (s *Service) CreateRole(ctx context.Context, code, name, description, scope string) (*model.Role, error) {
	if !roleCodePattern.MatchString(code) {
		return nil, errors.New("role code must match ^[a-z0-9][a-z0-9-]{1,31}$")
	}
	if scope == "" {
		scope = "global"
	}
	role := &model.Role{Code: code, Name: name, Description: description, Scope: scope}
	if err := s.r.RoleCreate(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) UpdateRole(ctx context.Context, id uint64, name, description, scope string) error {
	role, err := s.r.RoleGetByID(id)
	if err != nil {
		return err
	}
	role.Name = name
	role.Description = description
	if scope != "" {
		role.Scope = scope
	}
	return s.r.RoleUpdate(role)
}

func (s *Service) DeleteRole(ctx context.Context, id uint64) error {
	role, err := s.r.RoleGetByID(id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return errors.New("system role cannot be deleted")
	}
	return s.r.RoleDelete(id)
}

// GrantPerms 替换角色权限集，并使受影响用户的权限缓存失效。
func (s *Service) GrantPerms(ctx context.Context, roleID uint64, permCodes []string) error {
	if _, err := s.r.RoleGetByID(roleID); err != nil {
		return err
	}
	var permIDs []uint64
	for _, code := range permCodes {
		p, err := s.r.PermissionGetByCode(code)
		if err != nil {
			return errors.New("unknown permission: " + code)
		}
		permIDs = append(permIDs, p.ID)
	}
	if err := s.r.RolePermissionReplace(roleID, permIDs); err != nil {
		return err
	}
	// 失效持有该角色用户的权限缓存
	var userIDs []uint64
	if err := s.r.DB.Model(&model.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &userIDs).Error; err == nil {
		for _, uid := range userIDs {
			s.InvalidatePermCache(ctx, uid)
		}
	}
	return nil
}

// ---------- 列表透传 ----------

func (s *Service) UserList(f repository.UserFilter) ([]model.User, int64, error) {
	return s.r.UserList(f)
}

func (s *Service) UserByID(id uint64) (*model.User, error) {
	return s.r.UserGetByID(id)
}

func (s *Service) RoleList() ([]model.Role, error) {
	return s.r.RoleList()
}

func (s *Service) RolePermCodes(roleID uint64) ([]string, error) {
	return s.r.RolePermCodes([]uint64{roleID})
}

func (s *Service) PermissionList() ([]model.Permission, error) {
	return s.r.PermissionList()
}
