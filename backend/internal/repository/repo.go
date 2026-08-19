// Package repository 提供 GORM 数据访问层。
// 铁律：所有租户相关查询必须链式调用 ScopeByOrg/ScopeByProject（见 scope.go），
// 资源 ID 永不做权限判断依据。
package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type Repos struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Repos {
	return &Repos{DB: db}
}

// ---------- User ----------

func (r *Repos) UserCreate(u *model.User) error {
	return r.DB.Create(u).Error
}

func (r *Repos) UserGetByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *Repos) UserGetByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *Repos) UserGetByID(id uint64) (*model.User, error) {
	var u model.User
	err := r.DB.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

type UserFilter struct {
	Keyword string
	Page    int
	Size    int
}

func (r *Repos) UserList(f UserFilter) ([]model.User, int64, error) {
	q := r.DB.Model(&model.User{})
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.User
	err := q.Order("id DESC").Offset((f.Page - 1) * f.Size).Limit(f.Size).Find(&items).Error
	return items, total, err
}

func (r *Repos) UserUpdate(u *model.User) error {
	return r.DB.Model(u).Select("nickname", "avatar_url", "status", "password_hash").Updates(u).Error
}

func (r *Repos) UserSoftDelete(id uint64) error {
	return r.DB.Delete(&model.User{}, id).Error
}

// ---------- User Role ----------

func (r *Repos) UserRoleReplace(userID uint64, roleIDs []uint64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&model.UserRole{UserID: userID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repos) UserRoleCodes(userID uint64) ([]string, error) {
	var codes []string
	err := r.DB.Model(&model.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Pluck("roles.code", &codes).Error
	return codes, err
}

func (r *Repos) UserRoleNames(userID uint64) ([]string, error) {
	var names []string
	err := r.DB.Model(&model.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Order("roles.id ASC").
		Pluck("roles.name", &names).Error
	return names, err
}

func (r *Repos) UserRoleIDs(userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.DB.Model(&model.UserRole{}).Where("user_id = ?", userID).Pluck("role_id", &ids).Error
	return ids, err
}

// ---------- Role ----------

func (r *Repos) RoleList() ([]model.Role, error) {
	var items []model.Role
	err := r.DB.Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repos) RoleGetByCode(code string) (*model.Role, error) {
	var role model.Role
	err := r.DB.Where("code = ?", code).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &role, err
}

func (r *Repos) RoleGetByID(id uint64) (*model.Role, error) {
	var role model.Role
	err := r.DB.First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &role, err
}

func (r *Repos) RoleCreate(role *model.Role) error {
	return r.DB.Create(role).Error
}

func (r *Repos) RoleUpdate(role *model.Role) error {
	return r.DB.Model(role).Select("name", "description", "scope").Updates(role).Error
}

func (r *Repos) RoleDelete(id uint64) error {
	return r.DB.Delete(&model.Role{}, id).Error
}

func (r *Repos) RolePermissionReplace(roleID uint64, permIDs []uint64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		for _, pid := range permIDs {
			if err := tx.Create(&model.RolePermission{RoleID: roleID, PermissionID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repos) RolePermCodes(roleIDs []uint64) ([]string, error) {
	var codes []string
	err := r.DB.Model(&model.RolePermission{}).
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id IN ?", roleIDs).
		Distinct().
		Pluck("permissions.code", &codes).Error
	return codes, err
}

// ---------- Permission ----------

func (r *Repos) PermissionList() ([]model.Permission, error) {
	var items []model.Permission
	err := r.DB.Order("module ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repos) PermissionGetByCode(code string) (*model.Permission, error) {
	var p model.Permission
	err := r.DB.Where("code = ?", code).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// ---------- 日志 ----------

func (r *Repos) LoginLogCreate(l *model.LoginLog) error {
	return r.DB.Create(l).Error
}

func (r *Repos) AuditLogCreate(l *model.AuditLog) error {
	return r.DB.Create(l).Error
}
