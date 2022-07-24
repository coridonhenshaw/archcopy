package streamfile

import (
	"sync"
	"time"

	"github.com/DataDog/zstd"
)

type StreamfileSendStruct struct {
	WriteFunc  func(*CompressedBlockStruct) error
	CancelFunc func()

	EnableCompression bool
	ForceCompression  bool

	CompressionQueue      chan *[]byte
	CompressionCancelChan chan bool

	SendQueue        chan *CompressedBlockStruct
	SendCancelChan   chan bool
	SendCompleteChan chan bool

	StoredError error
	Mutex       sync.Mutex

	CompressH *HistogramStruct
	WriteH    *HistogramStruct
	CompressL *HistogramStruct

	Wait1 time.Duration
	Wait2 time.Duration
}

type CompressedBlockStruct struct {
	DecompressedSize uint64
	Data             []byte
	Compressed       bool
	Zero             int64
}

type HistogramStruct struct {
	Bins []int
}

func GetHistogram(Bins int) *HistogramStruct {
	var o HistogramStruct
	o.Bins = make([]int, Bins)
	return &o
}

const CompressionQueueSize = 192
const SendQueueSize = 256

func (o *StreamfileSendStruct) Initialize() {

	o.EnableCompression = true
	o.ForceCompression = false

	o.CompressionQueue = make(chan *[]byte, CompressionQueueSize)
	o.CompressionCancelChan = make(chan bool, 8)

	o.SendQueue = make(chan *CompressedBlockStruct, SendQueueSize)
	o.SendCompleteChan = make(chan bool)
	o.SendCancelChan = make(chan bool, 8)

	go o.CompressThread()
	go o.SendThread()

	o.CompressH = GetHistogram(CompressionQueueSize + 1)
	o.WriteH = GetHistogram(SendQueueSize + 1)
	o.CompressL = GetHistogram(42)
}

func (o *StreamfileSendStruct) Cancel(err error) {
	o.Mutex.Lock()
	o.StoredError = err
	o.Mutex.Unlock()

	o.SendCancelChan <- true
	o.CompressionCancelChan <- true

	for {
		select {
		case <-o.SendQueue:
		case <-o.CompressionQueue:
		default:
			break
		}
	}
}

func (o *StreamfileSendStruct) CompressThread() {
	var err error
Outer:
	for {
		var In *[]byte
		select {
		case In = <-o.CompressionQueue:
			if In == nil || len(*In) == 0 {
				o.SendQueue <- nil
				break Outer
			}
		case <-o.CompressionCancelChan:
			break Outer
		}

		err = ZeroFilter(In, func(Start int, End int, Sparse bool) error {
			var ierr error
			Out := CompressedBlockStruct{DecompressedSize: uint64(End - Start)}
			if Sparse {
				Out.Zero = int64(End - Start)
			} else {
				SendQueueLength := len(o.SendQueue)
				CompressionQueueLength := len(o.CompressionQueue)
				if o.ForceCompression || (o.EnableCompression && (SendQueueLength > 1 || CompressionQueueLength == 0)) {
					Out.Compressed = true
					var Level int = 40
					if CompressionQueueLength > 0 {
						Level = int((float64(SendQueueLength) / float64(SendQueueSize) * 40.0))
					}

					o.CompressL.Bins[Level+1]++

					Level -= 20

					if Level < -20 {
						Level = -20
					} else if Level > 20 {
						Level = 20
					}
					Out.Data, ierr = zstd.CompressLevel(nil, (*In)[Start:End], Level-20)
				} else {
					o.CompressL.Bins[0]++
					Out.Compressed = false
					Out.Data = (*In)[Start:End]
				}
			}
			select {
			case o.SendQueue <- &Out:
			case <-o.CompressionCancelChan:
			}

			return ierr
		})

		if err != nil {
			o.Cancel(err)
			break
		}
	}
}

func (o *StreamfileSendStruct) SendThread() {
	var err error

Outer:
	for err == nil {
		var In *CompressedBlockStruct
		tStart := time.Now()

		select {
		case In = <-o.SendQueue:
		case <-o.SendCancelChan:
			break Outer
		}

		o.Wait1 += time.Now().Sub(tStart)

		if In == nil || (len(In.Data) == 0 && In.Zero == 0) {
			break
		}
		o.CompressH.Bins[len(o.CompressionQueue)]++
		o.WriteH.Bins[len(o.SendQueue)]++

		tStart = time.Now()

		err = o.WriteFunc(In)

		o.Wait2 += time.Now().Sub(tStart)
	}

	if err != nil {
		o.Cancel(err)
	}

	o.SendCompleteChan <- true
}

func (o *StreamfileSendStruct) In(Block *[]byte) error {
	var err error

	o.Mutex.Lock()
	err = o.StoredError
	o.Mutex.Unlock()

	if err != nil {
		return err
	}

	o.CompressionQueue <- Block
	return nil
}

func (o *StreamfileSendStruct) Await() error {

	var err error

	o.Mutex.Lock()
	err = o.StoredError
	o.Mutex.Unlock()

	if err != nil {
		return err
	}

	o.CompressionQueue <- nil
	<-o.SendCompleteChan

	o.Mutex.Lock()
	err = o.StoredError
	o.Mutex.Unlock()

	// fmt.Print("\n\n\n")
	// fmt.Println("Level:", o.CompressL.Bins)
	// fmt.Println("Input:", o.CompressH.Bins)
	// fmt.Println(o.CompressH.Stats())
	// fmt.Println("Send:", o.WriteH.Bins)
	// fmt.Println(o.WriteH.Stats())
	// fmt.Println(o.Wait1, o.Wait2)
	// fmt.Print("\n\n\n")

	return err
}

func (o *HistogramStruct) Stats() (Min int, Max int) {
	var Total int
	for _, v := range o.Bins {
		Total += v
	}

	Min = int((float64(o.Bins[0]) / float64(Total)) * float64(100.0))
	Max = int((float64(o.Bins[len(o.Bins)-1]) / float64(Total)) * float64(100.0))

	return
}
