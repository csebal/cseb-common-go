package websocket

import (
	"time"
)

type Config struct {
	Address     string        `json:"address"`
	ConnTimeout time.Duration `json:"connTimeout"`
	SendTimeout time.Duration `json:"sendTimeout"`
}

func DefaultConfig() *Config {
	return &Config{
		ConnTimeout: 30 * time.Second,
		SendTimeout: 30 * time.Second,
	}
}

func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

func (c *Config) WithConnectTimeout(timeout time.Duration) *Config {
	c.ConnTimeout = timeout
	return c
}

func (c *Config) WithSendTimeout(timeout time.Duration) *Config {
	c.SendTimeout = timeout
	return c
}
