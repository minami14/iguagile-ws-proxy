package proxy

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net"

	"github.com/gorilla/websocket"
)

type tcpClient struct {
	conn net.Conn
	buf  []byte
}

func dial(address string) (*tcpClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return &tcpClient{
		conn: conn,
		buf:  make([]byte, 2),
	}, nil
}

// Read reads data from the connection.
func (t *tcpClient) Read(b []byte) (int, error) {
	if _, err := t.conn.Read(t.buf); err != nil {
		return 0, err
	}

	size := int(binary.LittleEndian.Uint16(t.buf))
	receivedSizeSum := 0
	for receivedSizeSum < size {
		receivedSize, err := t.conn.Read(b[receivedSizeSum:size])
		if err != nil {
			return 0, err
		}
		receivedSizeSum += receivedSize
	}

	return size, nil
}

// Write writes data to the connection.
func (t *tcpClient) Write(b []byte) (int, error) {
	size := len(b)
	sizeByte := make([]byte, 2, size+2)
	binary.LittleEndian.PutUint16(sizeByte, uint16(size))
	b = append(sizeByte, b...)
	if _, err := t.conn.Write(b); err != nil {
		return 0, err
	}
	return size, nil
}

func (t *tcpClient) verify(room Room) error {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint32(buf, uint32(room.RoomId))
	if _, err := t.Write(buf[:4]); err != nil {
		return err
	}

	if _, err := t.Write([]byte(room.ApplicationName)); err != nil {
		return err
	}

	if _, err := t.Write([]byte(room.Version)); err != nil {
		return err
	}

	if _, err := t.Write([]byte(room.Password)); err != nil {
		return err
	}

	if room.Token != "" {
		n, err := base64.StdEncoding.Decode(buf, []byte(room.Token))
		if err != nil {
			return err
		}

		if _, err := t.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

func (t *tcpClient) readStart(ctx context.Context, ws *websocket.Conn) error {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := t.Read(buf)
			if err != nil {
				return err
			}

			if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				return err
			}
		}
	}
}

var errConnectionClosed = errors.New("websocket connection closed")

func (t *tcpClient) writeStart(ctx context.Context, ws *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			mt, message, err := ws.ReadMessage()
			if err != nil {
				return err
			}

			if mt == websocket.TextMessage || mt == websocket.BinaryMessage {
				if _, err := t.Write(message); err != nil {
					return err
				}

				continue
			}

			switch mt {
			case websocket.PingMessage:
				if err := ws.WriteMessage(websocket.PongMessage, message); err != nil {
					return err
				}
			case websocket.CloseMessage:
				return errConnectionClosed
			}
		}
	}
}
