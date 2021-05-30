package lokas

import (
	"bufio"
	"time"
)

// Content context for create a dialer or a listener
type Context struct {
	SessionCreator        func(conn IConn) ISession
	Splitter              bufio.SplitFunc      // packet splitter
	IPChecker             func(ip string) bool // check if an accepted connection is allowed
	IdleTimeAfterOpen     time.Duration        // idle time when open, conn will be closed if not activated after this time
	ReadBufferSize        int                  // buffer size for reading
	WriteBufferSize       int                  // buffer size for writing
	UseNoneBlockingChan   bool                 // use none blocking chan
	ChanSize              int                  // chan size for bufferring
	MaxMessageSize        int                  // max message size for a single packet
	MergedWriteBufferSize int                  // buffer size for merged write
	DisableMergedWrite    bool                 // disable merge multiple message to a single net.Write
	EnableStatistics      bool                 // enable statistics of packets send and recv
	Extra                 interface{}          // used for special cases when custom data is needed
	LongPacketPicker      LongPacketPicker     // check and pick long packet when recv
	LongPacketCreator     LongPacketCreator    // create long packet for send
	MaxPacketWriteLen     int                  // data size for long packet
}
