package websocket

import (
	"context"
	"go.uber.org/zap/zaptest"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestParams_WithLogger(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := DefaultParams().WithLogger(logger)
	if p.logger != logger {
		t.Errorf("Expected logger to be %v, got %v", logger, p.logger)
	}
}

func TestParams_WithConfig(t *testing.T) {
	config := DefaultConfig().WithAddress("test1")
	p := DefaultParams().WithConfig(*config)
	if p.cfg.Address != "test1" {
		t.Errorf("Expected cfg.address to be test1, got %v", p.cfg.Address)
	}
}

func TestParams_WithConnFactory(t *testing.T) {
	var sf ConnFactoryFunc = func(ctx context.Context, u string) (ConnHandler, *http.Response, error) {
		return nil, nil, nil
	}
	p := DefaultParams().WithConnFactory(sf)
	if reflect.ValueOf(p.connFactory).Pointer() != reflect.ValueOf(sf).Pointer() {
		t.Fatalf("Failed to properly set Connection Factory")
	}
}

func TestParams_WithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := DefaultParams().WithContext(ctx)
	if p.ctx != ctx {
		t.Fatalf("failed to properly set Context")
	}
	cancel()
	select {
	case _, ok := <-p.ctx.Done():
		if !ok {
			return
		}
		t.Fatalf("context should be done, but it is not")
	case <-time.After(10 * time.Millisecond):
		t.Fatalf("context should be done, but it is not")
	}
}
