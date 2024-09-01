package websocket

import (
	"context"
	"errors"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"
)

type ClientTestContext struct {
	testAddr  string
	params    Params
	connErr   error
	msgRecvr  func(m SocketData) error
	msgSender chan SocketData
}

func NewClientTestContext() *ClientTestContext {
	c := &ClientTestContext{
		testAddr: "wss://localhost:3333",
	}
	c.params = *DefaultParams().
		WithConnFactory(c.mockConnect).
		WithConfig(*DefaultConfig().WithAddress("wss://localhost:3333"))

	return c
}

func (tc *ClientTestContext) mockConnect(ctx context.Context, addr string) (ConnHandler, *http.Response, error) {
	return NewMockConn(addr, tc.msgRecvr, tc.msgSender), nil, tc.connErr
}

func TestClient_NewClient(t *testing.T) {
	ct := NewClientTestContext()

	c := NewClient(ct.params)

	if c.params.cfg.Address != ct.params.cfg.Address {
		t.Fatalf("expected cfg to be set to client")
	}
	if c.connected {
		t.Fatalf("Connected client should not be connected")
	}
}

func TestClient_Connect_HappyPath(t *testing.T) {
	ct := NewClientTestContext()

	c := NewClient(ct.params)
	err := c.Connect()
	defer func(c *Client) {
		_ = c.Disconnect()
	}(c)

	if err != nil {
		t.Fatalf("expected err to be nil, but was: %v", err)
	}
	if !c.connected {
		t.Fatalf("client is not connected")
	}
}

func TestClient_Connect_Timeout(t *testing.T) {
	ct := NewClientTestContext()
	ct.params = *DefaultParams().
		WithConfig(*DefaultConfig().WithAddress(ct.testAddr).WithConnectTimeout(0 * time.Millisecond))

	c := NewClient(ct.params)
	e := c.Connect()
	defer func(c *Client) {
		_ = c.Disconnect()
	}(c)
	if errors.Is(e, ErrSocketTimeout) {
		ei := errors.Unwrap(e)
		if !ei.(net.Error).Timeout() {
			t.Fatalf("expected err to be Context deadline exceeded, but was: %v", e)
		}
	}

	if c.connected {
		t.Fatalf("client is connected, when it shouldn't be")
	}
}

func TestClient_Connect_Refused(t *testing.T) {
	ct := NewClientTestContext()
	ct.params = *DefaultParams().
		WithConfig(*DefaultConfig().WithAddress(ct.testAddr))
	c := NewClient(ct.params)
	err := c.Connect()
	defer func(c *Client) {
		_ = c.Disconnect()
	}(c)
	if errors.Is(err, ErrFailedToConnect) {
		ei := errors.Unwrap(err)
		if errors.Is(ei, syscall.ECONNREFUSED) {
			t.Fatalf("expected err to be Context deadline exceeded, but was: %v", err)
		}
	}
	if c.connected {
		t.Fatalf("client is connected, when it shouldn't be")
	}
}

func TestClient_Disconnect(t *testing.T) {
	ct := NewClientTestContext()

	c := NewClient(ct.params)
	_ = c.Connect()
	_ = c.Disconnect()

	if c.connected {
		t.Fatalf("client is connected, when it shouldn't be")
	}

	if c.conn != nil {
		t.Fatalf("client still holds a connection, when it's disconnected")
	}
}

func TestClient_Send_HappyPath(t *testing.T) {
	ct := NewClientTestContext()

	var r SocketData
	done := make(chan struct{})
	ct.msgRecvr = func(m SocketData) error {
		r = m
		close(done)
		return nil
	}

	c := NewClient(ct.params)
	_ = c.Connect()
	defer func(c *Client) {
		_ = c.Disconnect()
	}(c)

	tdm := "Hello World"

	err := c.Send([]byte(tdm))
	if err != nil {
		t.Fatalf("unexpected error sending data: %v", err)
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timed out waiting for MockRecvdMsg")
	}

	if r.err != nil {
		t.Fatalf("expected err to be nil, but was: %v", r.err)
	}

	if r.data == nil {
		t.Fatalf("MockRecvdMsg not sent")
	}

	if string(tdm) != string(r.data) {
		t.Fatalf("Incorrect MockRecvdMsg bytes")
	}
}

func TestClient_Send_NotConnected(t *testing.T) {
	ct := NewClientTestContext()

	c := NewClient(ct.params)

	tdm := "Hello World"

	err := c.Send([]byte(tdm))
	if !errors.Is(err, ErrInvalidSocketState) {
		t.Fatalf("unexpected error sending data: %v", err)
	}
}

func TestClient_Send_Timeout(t *testing.T) {
	ct := NewClientTestContext()
	ct.params = *ct.params.
		WithConfig(*ct.params.cfg.WithSendTimeout(1 * time.Millisecond))

	done := make(chan struct{})
	ct.msgRecvr = func(m SocketData) error {
		<-time.After(10 * time.Millisecond)
		close(done)
		return nil
	}

	c := NewClient(ct.params)
	_ = c.Connect()
	defer func(c *Client) {
		_ = c.Disconnect()
	}(c)

	tdm := "Hello World"

	err := c.Send([]byte(tdm))
	if !errors.Is(err, ErrSocketTimeout) {
		t.Fatalf("unexpected error sending data: %v", err)
	}
}

func TestClient_Listen(t *testing.T) {
	ct := NewClientTestContext()
	ct.msgSender = make(chan SocketData)

	c := NewClient(ct.params)
	_ = c.Connect()
	defer func(c *Client) {
		_ = c.Disconnect()
	}(c)

	go func(ch chan SocketData) {
		ch <- SocketData{
			msgType: 0,
			data:    []byte("Hello World"),
			err:     nil,
		}
	}(ct.msgSender)

	lch := c.Listen()
	select {
	case r := <-lch:
		if r.msgType != 0 || string(r.data) != "Hello World" {
			t.Fatalf("unexpected message: %v - %v", r.msgType, string(r.data))
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timed out waiting for message to arrive")
	}
}
