package service

import (
	"strings"
	"testing"

	"github.com/dxcloud/cloud-api/internal/model"
)

func TestOrgMatches(t *testing.T) {
	orgID := uint64(10)
	if !orgMatches(orgID, &orgID) {
		t.Fatal("same org should match")
	}
	if orgMatches(0, &orgID) {
		t.Fatal("default context should not match org resource")
	}
	if !orgMatches(0, nil) {
		t.Fatal("default context should match nil org")
	}
}

func TestAccessAllowed(t *testing.T) {
	owner := uint64(2)
	orgID := uint64(10)
	orgInst := &model.EcsInstance{OrgID: &orgID, OwnerID: owner}
	defaultInst := &model.EcsInstance{OrgID: nil, OwnerID: owner}
	if accessAllowed(AccessCtx{UserID: owner, Roles: []string{"user"}, OrgID: 0}, orgInst, false) {
		t.Fatal("default context should not allow org resource by ID")
	}
	if !accessAllowed(AccessCtx{UserID: owner, Roles: []string{"user"}, OrgID: orgID}, orgInst, false) {
		t.Fatal("same org owner should be allowed")
	}
	if !accessAllowed(AccessCtx{UserID: owner, Roles: []string{"user"}, OrgID: 0}, defaultInst, false) {
		t.Fatal("default owner should be allowed")
	}
}

func TestAppendPullLogLineCollapsesNoise(t *testing.T) {
	svc := &ImageService{
		pullLogs:  make(map[uint64]string),
		pullLines: make(map[uint64]string),
		noisySeen: make(map[uint64]map[string]bool),
	}
	for i := 0; i < 50; i++ {
		svc.appendPullLogLine(1, "Downloading")
	}
	svc.appendPullLogLine(1, "Pull complete")
	logs := svc.pullLogs[1]
	if strings.Count(logs, "Downloading") > 1 {
		t.Fatalf("repeated Downloading should be collapsed: %s", logs)
	}
	if !strings.Contains(logs, "后续重复进度已折叠") {
		t.Fatalf("expected collapse hint, got %s", logs)
	}
	if !strings.Contains(logs, "Pull complete") {
		t.Fatalf("expected real progress to remain, got %s", logs)
	}
}

func TestParseDockerProgress(t *testing.T) {
	percent, ok := parseDockerProgress("12.5MB/50MB")
	if !ok || percent != 25 {
		t.Fatalf("expected 25%%, got %v ok=%v", percent, ok)
	}
	if _, ok := parseDockerProgress("Downloading"); ok {
		t.Fatal("status-only line should not produce a percent")
	}
}
