package websocket

import (
	"context"
	"net/http"
)

type ConnReader interface {
	ReadMessage() (messageType int, data []byte, err error)
}

type ConnWriter interface {
	WriteMessage(messageType int, data []byte) error
}

type ConnCloser interface {
	Close() error
}

type ConnHandler interface {
	ConnReader
	ConnWriter
	ConnCloser
}

type ConnFactoryFunc func(ctx context.Context, u string) (ConnHandler, *http.Response, error)
