package service

import (
	"strings"
	"testing"

	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestResolvePullRefsIncludesFallbacks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.SystemSetting{}); err != nil {
		t.Fatal(err)
	}
	repos := repository.New(db)
	_ = repos.SettingsUpsert(&model.SystemSetting{Key: model.SettingRegion, Value: `"cn"`})
	_ = repos.SettingsUpsert(&model.SystemSetting{Key: model.SettingRegistryMirror, Value: `"hub.rat.dev"`})

	svc := NewSettingsService(repos, zap.NewNop())
	refs, mirrored := svc.ResolvePullRefs("nginx:latest")
	if !mirrored {
		t.Fatal("cn region should use mirrors")
	}
	if len(refs) < 2 {
		t.Fatalf("expected fallback refs, got %d: %v", len(refs), refs)
	}
	if !strings.HasPrefix(refs[0], "hub.rat.dev/library/nginx:") {
		t.Fatalf("primary should use configured mirror, got %s", refs[0])
	}
	found := false
	for _, ref := range refs {
		if strings.HasPrefix(ref, "docker.1ms.run/library/nginx:") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("fallback list should include docker.1ms.run: %v", refs)
	}
}
