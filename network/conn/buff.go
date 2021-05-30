package conn

import (
	"bytes"
)

type DataBuff struct {
	buff     *bytes.Buffer
	buffSize int
	enabled  bool
}

const MinMergedWriteBuffSize = 100 * 1024

func NewDataBuff(buffSize int, enabled bool) *DataBuff {
	if !enabled {
		return &DataBuff{enabled: enabled}
	}
	return &DataBuff{
		buff:     bytes.NewBuffer(make([]byte, buffSize+2*1024)), // buff cap is a little bigger than buffSize
		buffSize: buffSize,
	}
}

// GetData return merged buff from channel c
func (this *DataBuff) GetData(data []byte, c <-chan []byte) ([]byte, int) {
	if !this.enabled || len(c) == 0 {
		return data, 1
	}

	buff, buffSize, count := this.buff, this.buffSize, 0
	buff.Reset()
	for {
		count++
		buff.Write(data)
		if len(c) == 0 || buff.Len() >= buffSize {
			break
		}
		data = <-c
	}
	return buff.Bytes(), count
}

func (this *DataBuff) WriteData(data []byte, c <-chan []byte, spilt func(data []byte) ([][]byte, error), maxPacketWriteLen int, write func(rb []byte, count int) error) error {
	if spilt == nil || len(data) <= maxPacketWriteLen {
		rb, count := this.GetData(data, c)
		return write(rb, count)
	}

	splitDatas, err := spilt(data)
	if err != nil {
		return err
	}
	splitLen := len(splitDatas)
	if splitLen <= 0 {
		return nil
	}

	lastIdx := splitLen - 1
	if !this.enabled {
		for idx, splitData := range splitDatas {
			if idx == lastIdx {
				err = write(splitData, 1)
			} else {
				err = write(splitData, 0)
			}
			if err != nil {
				return err
			}
		}
	} else {
		buff, buffSize := this.buff, this.buffSize
		buff.Reset()

		needContinue := false
		for idx, splitData := range splitDatas {
			buff.Write(splitData)

			if buff.Len() >= buffSize {
				if idx == lastIdx {
					err = write(buff.Bytes(), 1)
				} else {
					err = write(buff.Bytes(), 0)
				}

				if err != nil {
					break
				}

				buff.Reset()
			} else {
				if idx == lastIdx {
					needContinue = true
				}
			}
		}
		if needContinue {
			rb, count := this.GetData(buff.Bytes(), c)
			return write(rb, count)
		}
	}

	return nil
}
