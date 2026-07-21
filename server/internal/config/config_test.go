package config

import "testing"

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("STAR_HTTP_PORT", "invalid")
	if _, err := LoadValidated(); err == nil {
		t.Fatal("Load() 应拒绝非法端口")
	}
}

func TestRequireProdCheck(t *testing.T) {
	t.Setenv("STAR_ENV", "prod")
	t.Setenv("STAR_HTTP_PORT", "8080")
	t.Setenv("STAR_JWT_SECRET", "short")
	t.Setenv("STAR_STATIC_CODE", "")
	cfg, err := LoadValidated()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := cfg.RequireProdCheck(); err == nil {
		t.Fatal("生产环境应拒绝短 JWT 密钥")
	}
}

func TestRoleForPhoneTrimsLists(t *testing.T) {
	t.Setenv("STAR_HTTP_PORT", "8080")
	t.Setenv("STAR_SALES_PHONES", " 13900000001,13900000002 ")
	t.Setenv("STAR_SUPPLIER_PHONES", " 13700000001 ")
	cfg, err := LoadValidated()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.RoleForPhone("13900000001"); got != "sales" {
		t.Fatalf("RoleForPhone() = %q, want sales", got)
	}
	if got := cfg.RoleForPhone("13700000001"); got != "supplier" {
		t.Fatalf("RoleForPhone() = %q, want supplier", got)
	}
}

func TestCanSeedRequiresAllowListOrForce(t *testing.T) {
	cfg := &Config{DBName: "star", SeedAllowList: []string{"star_test"}}
	if cfg.CanSeed() {
		t.Fatal("未显式允许时不应对 star 执行破坏性 seed")
	}
	cfg.SeedForce = true
	if !cfg.CanSeed() {
		t.Fatal("显式 force 后应允许 seed")
	}
}
