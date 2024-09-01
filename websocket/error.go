package websocket

type SocketError struct {
	message string
	inner   error
}

func NewSocketError(msg string) *SocketError {
	return &SocketError{
		message: msg,
	}
}

func (e *SocketError) WithInnerError(err error) *SocketError {
	e.inner = err
	return e
}

func (e *SocketError) Error() string {
	return e.message
}

var ErrFailedToConnect = NewSocketError("Failed to connect")
var ErrUnexpectedSocketError = NewSocketError("Unexpected socket error")
var ErrInvalidSocketState = NewSocketError("Invalid socket state")
var ErrSocketTimeout = NewSocketError("Socket timed out")
