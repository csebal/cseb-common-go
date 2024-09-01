package websocket

type SocketData struct {
	msgType int
	data    []byte
	err     error
}
