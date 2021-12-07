package conn

import (
	"errors"
	"github.com/nomos/go-lokas"
	"go.uber.org/zap"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/internal/common"
	"github.com/nomos/go-lokas/network/internal/hub"
)

// wrap websocket.IConn to adopt net.IConn
type WsConnWrapper struct {
	*websocket.Conn
}

type WsPumper struct {
	wsConn              *websocket.Conn
	done                chan struct{}
	mergedWriteBuffSize int
	disableMergedWrite  bool

	longPacketPicker  lokas.LongPacketPicker
	longPacketCreator lokas.LongPacketCreator
	maxPacketWriteLen int
	longPacketData    []byte
}

func NewWsConn(c *websocket.Conn, context *lokas.Context, hub *hub.Hub) *Conn {
	done := make(chan struct{})
	msgChan := NewMessageChan(context.UseNoneBlockingChan, context.ChanSize, done)
	conn := newConn(&WsConnWrapper{Conn: c}, context, msgChan, hub, done)
	mergedWriteBuffSize := MinMergedWriteBuffSize
	if context.MergedWriteBufferSize > mergedWriteBuffSize {
		mergedWriteBuffSize = context.MergedWriteBufferSize
	}
	if context.MaxMessageSize > 0 && mergedWriteBuffSize > context.MaxMessageSize {
		mergedWriteBuffSize = context.MaxMessageSize
	}
	conn.ioPumper = &WsPumper{
		wsConn:              c,
		done:                done,
		mergedWriteBuffSize: mergedWriteBuffSize,
		disableMergedWrite:  context.DisableMergedWrite,
		longPacketPicker:    context.LongPacketPicker,
		longPacketCreator:   context.LongPacketCreator,
		maxPacketWriteLen:   context.MaxPacketWriteLen,
	}
	return conn
}

func (this *WsConnWrapper) Read(b []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (this *WsConnWrapper) Write(data []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (this *WsConnWrapper) SetDeadline(t time.Time) error {
	return errors.New("not implemented")
}

func (this *WsPumper) readPump(conn *Conn) {
	wsConn := this.wsConn
	context := conn.context
	stat := conn.stat
	if context.MaxMessageSize > 0 {
		wsConn.SetReadLimit(int64(context.MaxMessageSize))
	}

	//wsConn.SetReadDeadline(time.Now().Add(common.PongWait)) // client must send ping, or conn will be closed after PongWait
	//wsConn.SetPongHandler(func(s string) error {
	//	wsConn.SetReadDeadline(time.Now().Add(common.PongWait))
	//	return nil
	//})

	for {
		_, data, err := wsConn.ReadMessage()
		if err != nil {
			log.Warn("read error",
				zap.String("err", err.Error()),
			)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
				log.Info("wsserver read error",
					zap.String("err", err.Error()),
				)
			}
			break
		}

		if this.longPacketPicker != nil {
			data = this.readLongPacket(conn, data)
		}

		if data != nil {
			conn.Session.OnRecv(conn, data)
		}

		if stat != nil {
			stat.AddRecvStat(len(data), 1)
		}
	}
	conn.Close() // finish writePump
}

func (this *WsPumper) readLongPacket(conn *Conn, data []byte) []byte {
	isLongPacket, idx, packetData := this.longPacketPicker(data)
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

func (this *WsPumper) writePump(conn *Conn) {
	wsConn := this.wsConn
	pingTicker := time.NewTicker(common.PingPeriod)
	outChan := conn.msgChan.GetOutChan()
	buff := NewDataBuff(this.mergedWriteBuffSize, !this.disableMergedWrite)
loop:
	for {
		select {
		case data := <-outChan:
			// if data == nil { // graceful close indicator
			// 	break loop
			// }
			err := this.write(conn, data, buff, outChan)
			if err != nil {
				log.Info("wsserver write error: %s",
					zap.String("err", err.Error()),
				)
				break loop
			}
		case <-pingTicker.C:
			wsConn.SetWriteDeadline(time.Now().Add(common.WriteWait))
			err := wsConn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Info("wsserver send ping error: %s",
					zap.String("err", err.Error()),
				)
				break loop
			}
		case <-this.done:
			break loop
		}
	}
	pingTicker.Stop()
	conn.Close()
}

func (this *WsPumper) write(conn *Conn, data []byte, buff *DataBuff, outChan <-chan []byte) error {
	conn.SetWriteDeadline(time.Now().Add(common.WriteWait))

	if this.longPacketCreator != nil && this.maxPacketWriteLen > 0 {
		return buff.WriteData(data, outChan, this.spiltLongPacket, this.maxPacketWriteLen, func(rb []byte, count int) error {
			return this.doWrite(conn, rb, count, outChan)
		})
	} else {
		return buff.WriteData(data, outChan, nil, this.maxPacketWriteLen, func(rb []byte, count int) error {
			return this.doWrite(conn, rb, count, outChan)
		})
	}
}

func (this *WsPumper) doWrite(conn *Conn, rb []byte, count int, outChan <-chan []byte) error {
	wsConn := this.wsConn
	stat := conn.stat

	err := wsConn.WriteMessage(websocket.BinaryMessage, rb)
	if err != nil {
		return err
	}
	if stat != nil && count > 0 {
		stat.AddSendStat(len(rb), count)
		stat.SetSendChanItemCount(len(outChan))
	}

	return nil
}

func (this *WsPumper) spiltLongPacket(data []byte) ([][]byte, error) {
	packets := make([][]byte, 0)

	all := len(data)
	pos := 0
	idx := 1
	maxLen := this.maxPacketWriteLen

	for pos < all {
		slen := maxLen
		if all-pos <= maxLen {
			slen = all - pos
			idx = 0
		}

		cdata, err := this.longPacketCreator(data[pos:pos+slen], idx)

		if err != nil {
			return packets, err
		}

		packets = append(packets, cdata)

		idx++
		pos += slen
	}

	return packets, nil
}
