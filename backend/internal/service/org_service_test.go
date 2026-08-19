package service

import (
	"fmt"
	"testing"

	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestRepos(t *testing.T) *repository.Repos {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.EcsInstance{}, &model.ResourceQuota{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return repository.New(db)
}

var seq int

func ecsRow(orgID *uint64, ownerID uint64, cpu float64, mem int64) *model.EcsInstance {
	seq++
	return &model.EcsInstance{
		InstanceNo:   fmt.Sprintf("i-test-%d", seq),
		ContainerID:  fmt.Sprintf("ctr-test-%d", seq), // sqlite 下 '' 会撞唯一索引（MySQL 中 GORM 零值省略→NULL）
		OrgID:        orgID,
		OwnerID:      ownerID,
		CPU:          cpu,
		MemoryMB:     mem,
		Name:         "t",
		Image:        "busybox",
		DesiredState: "running",
	}
}

func mustCreate(t *testing.T, repos *repository.Repos, row *model.EcsInstance) {
	t.Helper()
	if err := repos.DB.Create(row).Error; err != nil {
		t.Fatalf("create row: %v", err)
	}
}

// 注意：AutoMigrate 对带 *uint64 的 OrgID 列使用 NULL；0 组织回退单租户维度（owner）。
func TestCheckEcsQuotaSingleTenantDefault(t *testing.T) {
	repos := newTestRepos(t)
	svc := NewQuotaService(repos)
	// 默认 5 实例 / 8 核 / 16GB
	if err := svc.CheckEcsQuota(0, 7, 1, 512); err != nil {
		t.Fatalf("create #1 should pass: %v", err)
	}
	for i := 0; i < 4; i++ {
		mustCreate(t, repos, ecsRow(nil, 7, 1, 512))
	}
	// 已有 4 台 + 新 1 台 = 5 台，OK
	if err := svc.CheckEcsQuota(0, 7, 1, 512); err != nil {
		t.Fatalf("5th instance should pass: %v", err)
	}
	mustCreate(t, repos, ecsRow(nil, 7, 1, 512))
	// 第 6 台超实例数
	if err := svc.CheckEcsQuota(0, 7, 1, 512); err == nil {
		t.Fatal("6th instance should exceed default quota")
	}
}

func TestCheckEcsQuotaOrgOverrideAndOrgWide(t *testing.T) {
	repos := newTestRepos(t)
	svc := NewQuotaService(repos)
	orgID := uint64(10)
	_ = repos.QuotaUpsert(&model.ResourceQuota{OrgID: orgID, ResourceType: "ecs_count", LimitValue: 2})
	_ = repos.QuotaUpsert(&model.ResourceQuota{OrgID: orgID, ResourceType: "cpu", LimitValue: 4})
	_ = repos.QuotaUpsert(&model.ResourceQuota{OrgID: orgID, ResourceType: "memory", LimitValue: 1024})

	// 用户 A 在组织内建 1 台
	if err := svc.CheckEcsQuota(orgID, 100, 1, 512); err != nil {
		t.Fatalf("org create should pass: %v", err)
	}
	mustCreate(t, repos, ecsRow(&orgID, 100, 1, 512))

	// 用户 B 同组织再建 1 台：组织维度汇总 = 2 台，通过
	if err := svc.CheckEcsQuota(orgID, 200, 1, 512); err != nil {
		t.Fatalf("org-wide 2nd instance should pass: %v", err)
	}
	mustCreate(t, repos, ecsRow(&orgID, 200, 1, 512))

	// 第 3 台（无论谁建）超组织实例配额
	if err := svc.CheckEcsQuota(orgID, 300, 0.5, 128); err == nil {
		t.Fatal("3rd instance should exceed org quota (org-wide)")
	}
}

func TestCheckEcsQuotaCpuMemoryLimits(t *testing.T) {
	repos := newTestRepos(t)
	svc := NewQuotaService(repos)
	orgID := uint64(11)
	_ = repos.QuotaUpsert(&model.ResourceQuota{OrgID: orgID, ResourceType: "cpu", LimitValue: 2})
	_ = repos.QuotaUpsert(&model.ResourceQuota{OrgID: orgID, ResourceType: "memory", LimitValue: 1024})

	if err := svc.CheckEcsQuota(orgID, 1, 2, 1024); err != nil {
		t.Fatalf("exact quota should pass: %v", err)
	}
	if err := svc.CheckEcsQuota(orgID, 1, 2.1, 0); err == nil {
		t.Fatal("cpu 2.1 > 2 should fail")
	}
	if err := svc.CheckEcsQuota(orgID, 1, 1, 2048); err == nil {
		t.Fatal("mem 2048 > 1024 should fail")
	}
}

// 未配置配额项回退默认（不能因 0 值误拒）
func TestCheckEcsQuotaUnsetFallback(t *testing.T) {
	repos := newTestRepos(t)
	svc := NewQuotaService(repos)
	orgID := uint64(12)
	// 仅设置 ecs_count，cpu/memory 未配置 → 默认 8 核 / 16GB
	_ = repos.QuotaUpsert(&model.ResourceQuota{OrgID: orgID, ResourceType: "ecs_count", LimitValue: 10})
	if err := svc.CheckEcsQuota(orgID, 1, 4, 4096); err != nil {
		t.Fatalf("unset cpu/mem should fall back to defaults: %v", err)
	}
}
