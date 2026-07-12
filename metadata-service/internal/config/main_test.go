package config

import "testing"

// TestLoadConfigDefaults guards against malformed struct tags: LoadConfig must
// succeed with an unconfigured environment. A broken tag (e.g. a slice option
// that collides with the tag's comma separator) makes envconfig.Process fail
// for every startup, which is invisible until the binary is actually run.
func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed with default environment: %v", err)
	}
	if cfg.Port == "" {
		t.Errorf("expected a default Port, got empty")
	}
	if cfg.Raft == nil {
		t.Fatalf("expected Raft config to be populated")
	}
	if cfg.Raft.Enabled {
		t.Errorf("expected Raft to be disabled by default")
	}
}
