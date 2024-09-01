package websocket

type MockRecvdMsg struct {
	mt   int
	data []byte
	err  error
}

type MockConn struct {
	url        string
	gotMessage func(m SocketData) error
	readChan   <-chan SocketData
}

func NewMockConn(u string, gotMessage func(m SocketData) error, readCh <-chan SocketData) *MockConn {
	return &MockConn{
		url:        u,
		gotMessage: gotMessage,
		readChan:   readCh,
	}
}

func (m *MockConn) ReadMessage() (messageType int, data []byte, err error) {
	msg := <-m.readChan
	return msg.msgType, msg.data, msg.err
}

func (m *MockConn) WriteMessage(messageType int, data []byte) error {
	return m.gotMessage(SocketData{
		msgType: messageType,
		data:    data,
		err:     nil,
	})
}

func (*MockConn) Close() error {
	return nil
}
