package websocket

import (
	"context"
	"errors"
	ws "github.com/gorilla/websocket"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Connector interface {
	Connect() error
}

type Disconnector interface {
	Disconnect() error
}

type Listener interface {
	Listen() <-chan SocketData
}

type ClientHandler interface {
	Connector
	Disconnector
	Listener
}

type ClientFactoryFunc func(init Params) *Client

type Client struct {
	params     Params
	mmu        sync.RWMutex
	cmu        sync.RWMutex
	conn       ConnHandler
	rch        chan SocketData
	rchCancel  context.CancelFunc
	connecting bool
	connected  bool
}

func NewClient(init Params) *Client {
	if init.connFactory == nil {
		init.connFactory = func(ctx context.Context, u string) (ConnHandler, *http.Response, error) {
			dialer := &ws.Dialer{
				Proxy:            http.ProxyFromEnvironment,
				HandshakeTimeout: init.cfg.ConnTimeout,
			}
			ctx, cancel := context.WithTimeout(ctx, init.cfg.ConnTimeout)
			defer cancel()
			return dialer.DialContext(ctx, u, nil)
		}
	}
	return &Client{
		conn:      nil,
		params:    init,
		connected: false,
	}
}

func (c *Client) Connect() error {
	c.cmu.Lock()
	if c.connected || c.connecting {
		c.cmu.Unlock()
		return nil
	}
	c.mmu.Lock()
	c.cmu.Unlock()
	defer c.mmu.Unlock()
	if c.connecting {
		panic("we shouldn't be attempting to connect when the connecting flag is set")
	}
	if c.connected {
		return nil
	}
	c.connecting = true
	c.connected = false
	c.conn = nil
	conn, _, err := c.params.connFactory(c.params.ctx, c.params.cfg.Address)
	if err != nil {
		c.connecting = false
		return ErrFailedToConnect.WithInnerError(err)
	}
	c.conn = conn
	c.connected = true
	c.connecting = false
	c.rch = make(chan SocketData)
	var ctx context.Context
	ctx, c.rchCancel = context.WithCancel(c.params.ctx)
	go c.listenPump(ctx)
	return nil
}

func (c *Client) Disconnect() error {
	c.cmu.Lock()
	if !c.connected {
		c.cmu.Unlock()
		return nil
	}
	c.mmu.Lock()
	c.cmu.Unlock()
	defer c.mmu.Unlock()
	if !c.connected {
		return nil
	}
	c.connected = false
	var err error = nil
	if c.conn != nil {
		err = c.conn.Close()
		if err != nil {
			err = ErrUnexpectedSocketError.WithInnerError(err)
		}
	}
	c.rchCancel()
	c.conn = nil
	return err
}

func (c *Client) Send(message []byte) error {
	c.mmu.RLock()
	defer c.mmu.RUnlock()
	if !c.connected {
		return ErrInvalidSocketState.WithInnerError(errors.New("Client is disconnected"))
	}
	ch := make(chan error)
	defer close(ch)

	sc := c.params.ctx
	if c.params.cfg.SendTimeout > 0 {
		ctx, cancel := context.WithTimeout(c.params.ctx, c.params.cfg.SendTimeout)
		sc = ctx
		defer cancel()
	}

	go func(parentCtx context.Context, ch chan error) {
		err := c.conn.WriteMessage(0, message)
		if parentCtx.Err() == nil {
			if err != nil {
				err = ErrUnexpectedSocketError.WithInnerError(err)
			}
			ch <- err
		}
	}(sc, ch)

	select {
	case err := <-ch:
		return err
	case <-sc.Done():
		return ErrSocketTimeout.WithInnerError(sc.Err())
	}
	//c.params.logger.Debug("Queued MockRecvdMsg for sending", zap.ByteString("MockRecvdMsg", MockRecvdMsg))
}

func (c *Client) Listen() <-chan SocketData {
	c.mmu.RLock()
	defer c.mmu.RUnlock()
	return c.rch
}

func (c *Client) listenPump(ctx context.Context) {
	defer close(c.rch)
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sigs:
			return
		case <-ctx.Done():
			return
		default:
			if !c.connected {
				<-time.After(100 * time.Millisecond)
				continue
			}
			msgType, msgBody, err := c.conn.ReadMessage()
			if err != nil {
				c.rch <- SocketData{0, nil, ErrUnexpectedSocketError.WithInnerError(err)}
				//c.ctxLogger.Error("Socket Read Error", zap.Error(err))
			}
			select {
			case <-ctx.Done():
				return
			case <-sigs:
				return
			case c.rch <- SocketData{msgType, msgBody, nil}:
			}
		}
	}
}
