package pipeline

import (
	"strings"
	"testing"
)

func TestParseDefinitionValid(t *testing.T) {
	def := `
name: ci-cd
timeout: 30m
env:
  REGISTRY: dx-registry:5000
steps:
  - name: 拉取代码
    type: git
    url: https://github.com/example/repo.git
    branch: main
  - name: 构建镜像
    type: docker-build
    dockerfile: Dockerfile
    tags: [dx-registry:5000/app:v1]
  - name: 单测
    type: shell
    script: go test ./...
  - name: 等待健康
    type: wait-health
    url: http://localhost:8080/healthz
`
	p, err := ParseDefinition(def)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if p.Name != "ci-cd" || len(p.Steps) != 4 {
		t.Fatalf("unexpected def: %+v", p)
	}
	if p.Steps[1].Tags[0] != "dx-registry:5000/app:v1" {
		t.Fatalf("tags parse failed: %v", p.Steps[1].Tags)
	}
}

func TestParseDefinitionEmpty(t *testing.T) {
	if _, err := ParseDefinition(""); err == nil {
		t.Fatal("empty definition should fail")
	}
	if _, err := ParseDefinition("   "); err == nil {
		t.Fatal("blank definition should fail")
	}
}

func TestParseDefinitionNoSteps(t *testing.T) {
	if _, err := ParseDefinition("name: x\nsteps: []\n"); err == nil {
		t.Fatal("zero steps should fail")
	}
}

func TestParseDefinitionBadYAML(t *testing.T) {
	if _, err := ParseDefinition("name: [unclosed"); err == nil {
		t.Fatal("bad yaml should fail")
	}
}

func TestParseDefinitionStepValidation(t *testing.T) {
	cases := []struct {
		name string
		def  string
		want string
	}{
		{"missing name", "steps:\n  - type: shell\n    script: echo\n", "name"},
		{"unknown type", "steps:\n  - name: s\n    type: exec-privileged\n", "白名单"},
		{"shell no script", "steps:\n  - name: s\n    type: shell\n", "script"},
		{"git no url", "steps:\n  - name: s\n    type: git\n", "url"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDefinition(tc.def)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}
