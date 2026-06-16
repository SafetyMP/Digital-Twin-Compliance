package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Parallel()
	cfg := Load()
	if cfg.HTTPAddr != ":8091" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.PolicyDir != "policies/cedar" {
		t.Fatalf("PolicyDir = %q", cfg.PolicyDir)
	}
	if cfg.AuditTopic != "compliance.audit.pending" {
		t.Fatalf("AuditTopic = %q", cfg.AuditTopic)
	}
}
