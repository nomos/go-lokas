package conn

import (
	"bufio"
	"github.com/nomos/go-lokas"
	"go.uber.org/zap"
	"net"
	"time"

	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/internal/common"
	"github.com/nomos/go-lokas/network/internal/hub"
)

type TcpPumper struct {
	done                chan struct{}
	readBuffSize        int
	mergedWriteBuffSize int
	disableMergedWrite  bool

	longPacketPicker  lokas.LongPacketPicker
	longPacketCreator lokas.LongPacketCreator
	maxPacketWriteLen int
	longPacketData    []byte
}

const DefaultReadBuffSize = 8 * 1024
const DefaultWriteBuffSize = 16 * 1024
const ProtectLongPacketSize = 8 * 1024 * 1024

func NewTcpConn(c net.Conn, context *lokas.Context, hub *hub.Hub) *Conn {
	done := make(chan struct{})
	msgChan := NewMessageChan(context.UseNoneBlockingChan, context.ChanSize, done)
	conn := newConn(c, context, msgChan, hub, done)
	readBuffSize := DefaultReadBuffSize
	writeBuffSize := DefaultWriteBuffSize
	mergedWriteBuffSize := MinMergedWriteBuffSize
	if context.ReadBufferSize > 0 {
		readBuffSize = context.ReadBufferSize
	}
	if context.WriteBufferSize > 0 {
		writeBuffSize = context.WriteBufferSize
	}
	if context.MergedWriteBufferSize > mergedWriteBuffSize {
		mergedWriteBuffSize = context.MergedWriteBufferSize
	}
	c.(*net.TCPConn).SetReadBuffer(readBuffSize)
	c.(*net.TCPConn).SetWriteBuffer(writeBuffSize)
	conn.ioPumper = &TcpPumper{
		done:                done,
		readBuffSize:        readBuffSize,
		mergedWriteBuffSize: mergedWriteBuffSize,
		disableMergedWrite:  context.DisableMergedWrite,
		longPacketPicker:    context.LongPacketPicker,
		longPacketCreator:   context.LongPacketCreator,
		maxPacketWriteLen:   context.MaxPacketWriteLen,
	}
	return conn
}

func (this *TcpPumper) readPump(conn *Conn) {
	context := conn.context
	stat := conn.stat
	scanner := bufio.NewScanner(conn.Conn)
	scanner.Buffer(make([]byte, this.readBuffSize), this.readBuffSize)
	scanner.Split(context.Splitter)
	for {
		if ok := scanner.Scan(); ok {
			data := scanner.Bytes()

			if this.longPacketPicker != nil {
				data = this.readLongPacket(conn, data)
			}

			if data != nil {
				conn.Session.OnRecv(conn, data)
			}

			if stat != nil {
				stat.AddRecvStat(len(data), 1)
			}
		} else {
			break
		}
	}
	conn.Close() // finish writePump
}

func (this *TcpPumper) readLongPacket(conn *Conn, data []byte) []byte {
	isLongPacket, idx, packetData := this.longPacketPicker(data)
	if !isLongPacket {
		return data
	}

	// log.Info("readLongPacket,packetData = %v", packetData)
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

func (this *TcpPumper) writePump(conn *Conn) {
	tickerPing := time.NewTicker(common.PingPeriod)
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
				log.Info("tcpserver write error", zap.Error(err))
				break loop
			}
		case <-tickerPing.C:
		case <-this.done:
			break loop
		}
	}
	tickerPing.Stop()
	conn.Close()
}

func (this *TcpPumper) write(conn *Conn, data []byte, buff *DataBuff, outChan <-chan []byte) error {
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

func (this *TcpPumper) doWrite(conn *Conn, rb []byte, count int, outChan <-chan []byte) error {
	stat := conn.stat

	_, err := conn.Conn.Write(rb)
	if err != nil {
		return err
	}
	if stat != nil && count > 0 {
		stat.AddSendStat(len(rb), count)
		stat.SetSendChanItemCount(len(outChan))
	}

	return nil
}

func (this *TcpPumper) spiltLongPacket(data []byte) ([][]byte, error) {
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
