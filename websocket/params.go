package websocket

import (
	"context"
	"go.uber.org/zap"
)

type Params struct {
	cfg         Config
	logger      *zap.Logger
	ctx         context.Context
	connFactory ConnFactoryFunc
}

func DefaultParams() *Params {
	p := &Params{
		cfg:    *DefaultConfig(),
		ctx:    context.Background(),
		logger: zap.L(),
	}
	return p
}

func (p *Params) WithConfig(cfg Config) *Params {
	p.cfg = cfg
	return p
}

func (p *Params) WithLogger(logger *zap.Logger) *Params {
	p.logger = logger
	return p
}

func (p *Params) WithContext(ctx context.Context) *Params {
	p.ctx = ctx
	return p
}

func (p *Params) WithConnFactory(sf ConnFactoryFunc) *Params {
	p.connFactory = sf
	return p
}
