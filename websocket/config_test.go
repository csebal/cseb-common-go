package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestConfig_NewClientConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatalf("DefaultConfig() returned nil")
	}
	if cfg.Address != "" {
		t.Fatalf("DefaultConfig() returned non-empty address")
	}
}

func TestConfig_Unmarshal(t *testing.T) {
	type RootCFG struct {
		SCFG Config `json:"socket-config"`
	}

	data := []byte(`{"socket-config": {"address":"something", "connTimeout":1000000000, "sendTimeout":2000000000}}`)
	var cfg RootCFG

	err := json.Unmarshal(data, &cfg)
	if err != nil {
		t.Fatalf("marshalling returned non-nil err: %v", err)
	}
	if cfg.SCFG.Address != "something" || cfg.SCFG.ConnTimeout != time.Second*1 || cfg.SCFG.SendTimeout != time.Second*2 {
		t.Fatalf("unmarshalling returned invalid cfg, %v", cfg)
	}
}

func TestConfig_WithAddress(t *testing.T) {
	testAddress := "http://test.test.com"
	cfg := DefaultConfig().WithAddress(testAddress)
	if cfg.Address != testAddress {
		t.Fatalf("DefaultConfig() returned unexpected address")
	}
}

func TestConfig_WithConnectTimeout(t *testing.T) {
	testTimeout := 5 * time.Second
	cfg := DefaultConfig().WithConnectTimeout(testTimeout)
	if cfg.ConnTimeout != testTimeout {
		t.Fatalf("DefaultConfig() returned unexpected timeout")
	}
}

func TestConfig_WithSendTimeout(t *testing.T) {
	testTimeout := 5 * time.Second
	cfg := DefaultConfig().WithSendTimeout(testTimeout)
	if cfg.SendTimeout != testTimeout {
		t.Fatalf("DefaultConfig() returned unexpected timeout")
	}
}
