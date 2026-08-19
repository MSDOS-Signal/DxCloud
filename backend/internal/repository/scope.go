package repository

import "gorm.io/gorm"

// ScopeByOrg / ScopeByProject：租户隔离的铁律。
// 后续所有资源仓库（ecs/image/network/...）的查询必须链式调用，
// 从 SQL 层杜绝跨租户访问（IDOR 免疫）。

func ScopeByOrg(orgID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("org_id = ?", orgID)
	}
}

// ScopeByTenantOrNull：0 表示默认空间（org_id IS NULL），大于 0 表示指定组织。
// 租户资源永远按当前上下文过滤，避免默认空间看到其它组织数据。
func ScopeByTenantOrNull(orgID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgID == 0 {
			return db.Where("org_id IS NULL")
		}
		return db.Where("org_id = ?", orgID)
	}
}

func ScopeByProject(projectID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("project_id = ?", projectID)
	}
}

// ScopeByOwner 资源属主过滤（普通用户仅能操作自己创建的资源）。
func ScopeByOwner(ownerID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("owner_id = ?", ownerID)
	}
}
