package ws

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/promise"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"

	"strings"

	"github.com/gorilla/websocket"
)

const (
	HeaderSize = 8
	ProtectLongPacketSize = 4 * 1024 * 1024
)

type WebSocket struct {
	*websocket.Conn
	client         *WsClient
	writeChan      chan []byte
	wg             sync.WaitGroup
	closeOnce      sync.Once
	longPacketData []byte
	done           chan struct{}
	once           sync.Once
	closing        bool
}

func NewWebSocket(url string,client *WsClient) (*WebSocket, error) {
	ret := &WebSocket{
		Conn:      nil,
		writeChan: make(chan []byte),
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	ret.Conn = conn
	ret.client = client
	ret.ServeIO()
	return ret, nil
}

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 1024*1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024*1024,
	WriteBufferSize: 1024*1024*1024,
	CheckOrigin: func(r *http.Request) bool {
		return strings.HasPrefix(r.RemoteAddr, "127.0.0.1") || r.Header["Origin"][0] == r.Host
	},
}

func (this *WebSocket) ServeIO() {
	this.wg.Add(2)
	this.done = make(chan struct{})
	go func() {
		this.writePump()
		this.wg.Done()
	}()

	go func() {
		this.client.OnOpen(this.client.conn)
		this.readPump()
		this.client.OnClose(this.client.conn)
		this.wg.Done()
	}()
}

func (this *WebSocket) readPump() {
	defer func() {
		this.Conn.Close()
	}()

	this.Conn.SetReadLimit(maxMessageSize)
	this.Conn.SetReadDeadline(time.Now().Add(pongWait))
	this.Conn.SetPongHandler(func(string) error {
		this.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-this.done :
			return
		default:
			_, message, err := this.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Error("error: %v", zap.Error(err))
				}
				return
			}
			data := this.readLongPacket(message)
			this.client.OnRecvData(data)
		}
	}
}

func (this *WebSocket) readLongPacket(data []byte) []byte {
	isLongPacket, idx, packetData := protocol.PickBinaryLongPacket(data)
	if !isLongPacket {
		return data
	}

	this.longPacketData = append(this.longPacketData, packetData...)
	if idx == 0 {
		data := this.longPacketData[:]
		this.longPacketData = nil
		return data
	}

	//protect too long
	if len(this.longPacketData) > ProtectLongPacketSize {
		log.Error("protect too long")
		this.longPacketData = nil
	}
	return nil
}

func (this *WebSocket) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		this.Conn.Close()
	}()

	for {
		select {
		case <-this.done :
			return
		case res, ok := <-this.writeChan:
			this.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				this.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//data := make([]byte, len(res))
			//copy(data, res)
			//log.Warn("send res",len(data))
			//err := this.Conn.WriteMessage(websocket.BinaryMessage,data)
			//if err != nil {
			//	return
			//}

			w, err := this.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(res)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			this.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := this.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (this *WebSocket) Close() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		if this.closing {
			for {
				time.Sleep(time.Millisecond*50)
				if this.closing == false {
					resolve(nil)
					return
				}
			}
		} else {
			if this.done!=nil {
				this.done <- struct{}{}
				close(this.done)
			}
			this.wg.Wait()
			resolve(nil)
		}
	})
}

