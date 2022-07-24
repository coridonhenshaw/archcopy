package streamfile

import (
	"errors"
	"sync"
	"time"

	"github.com/DataDog/zstd"
	"proto.local/archcopyRPC"
	pb "proto.local/archcopyRPC"
)

type StreamfileReceiveStruct struct {
	PostDecompFunc func(dst *[]byte) error
	WriteFunc      func(dst *[]byte) error
	CancelFunc     func()

	Valid                  bool
	DecompressQueue        chan *archcopyRPC.File
	DecompressCompleteChan chan error
	DecompressCancelChan   chan bool
	WriteQueue             chan *[]byte
	WriteCompleteChan      chan bool
	WriteCancelChan        chan bool
	BytesRcvdCompressed    int
	BytesRcvdDecompressed  int
	Mutex                  sync.Mutex

	StoredError error

	DecompressH *HistogramStruct
	WriteH      *HistogramStruct

	Wait1 time.Duration
	Wait2 time.Duration
	Wait3 time.Duration
	Wait4 time.Duration
}

const DecompressQueueSize = 256
const WriteQueueSize = 64

func (o *StreamfileReceiveStruct) Initialize() {
	o.Valid = true
	o.DecompressQueue = make(chan *archcopyRPC.File, DecompressQueueSize)
	o.DecompressCancelChan = make(chan bool, 8)
	o.WriteQueue = make(chan *[]byte, WriteQueueSize)
	o.WriteCompleteChan = make(chan bool)
	o.WriteCancelChan = make(chan bool, 8)

	o.DecompressH = GetHistogram(DecompressQueueSize + 1)
	o.WriteH = GetHistogram(WriteQueueSize + 1)

	go o.WriteChunk()
	go o.DecompressChunk()
}

func (o *StreamfileReceiveStruct) Cancel(err error) {
	o.Mutex.Lock()
	o.StoredError = err
	o.Valid = false
	o.Mutex.Unlock()

	o.DecompressCancelChan <- true
	o.WriteCancelChan <- true
	if o.CancelFunc != nil {
		o.CancelFunc()
	}

	for {
		select {
		case <-o.DecompressQueue:
		case <-o.WriteQueue:
		default:
			break
		}
	}

}

func (o *StreamfileReceiveStruct) In(in *pb.File) error {
	o.Mutex.Lock()
	Valid := o.Valid
	o.Mutex.Unlock()

	if !Valid {
		if o.StoredError == nil {
			return errors.New("Not initialized")
		} else {
			return o.StoredError
		}
	}

	o.DecompressH.Bins[len(o.DecompressQueue)]++
	tStart := time.Now()
	o.DecompressQueue <- in
	o.Wait1 += time.Now().Sub(tStart)

	return nil
}

func (o *StreamfileReceiveStruct) DecompressChunk() {
	var err error

	var OutBuffer []byte

Outer:
	for err == nil {

		select {
		case Chunk := <-o.DecompressQueue:
			if Chunk == nil {
				break Outer
			}

			var dst []byte
			if Chunk.Zero > 0 {
				dst = make([]byte, Chunk.Zero)
			} else if Chunk.Compressed == true {
				buf := make([]byte, 65536*8)
				dst, err = zstd.Decompress(buf, Chunk.Data)
				if err != nil {
					o.Cancel(err)
					break Outer
				}
			} else {
				dst = Chunk.Data
			}

			if o.PostDecompFunc != nil {
				o.PostDecompFunc(&dst)
			}

			OutBuffer = append(OutBuffer, dst...)

			sz := len(OutBuffer)
			if sz > (65536 * 8) {
				Scratch := make([]byte, 65536*8)
				copy(Scratch, OutBuffer[0:65536*8])
				o.WriteH.Bins[len(o.WriteQueue)]++
				tStart := time.Now()
				o.WriteQueue <- &Scratch
				o.Wait2 += time.Now().Sub(tStart)
				OutBuffer = OutBuffer[65536*8:]
			}

			o.BytesRcvdCompressed += len(Chunk.Data)
			o.BytesRcvdDecompressed += len(dst)
		case <-o.DecompressCancelChan:
			err = errors.New("Cancelled")
			break Outer
		}
	}
	if err == nil {
		o.WriteQueue <- &OutBuffer
		o.WriteQueue <- nil
	}
}

func (o *StreamfileReceiveStruct) WriteChunk() {
	var err error
	var Block *[]byte

Outer:
	for err == nil {
		Timeout := time.NewTimer(1 * time.Minute)
		TimeoutChan := Timeout.C
		tStart := time.Now()
		select {
		case Block = <-o.WriteQueue:
			if Block == nil || len(*Block) == 0 {
				break Outer
			}
		case <-o.WriteCancelChan:
			break Outer
		case <-TimeoutChan:
			o.Cancel(errors.New("Timeout"))
		}
		Timeout.Stop()

		o.Wait3 += time.Now().Sub(tStart)

		tStart = time.Now()
		err = o.WriteFunc(Block)
		o.Wait4 += time.Now().Sub(tStart)
	}
	if err != nil {
		o.Cancel(err)
	}

	o.WriteCompleteChan <- true
}

func (o *StreamfileReceiveStruct) Await() (int, int, error) {

	// fmt.Println("\nDecomp:", o.DecompressH.Bins)
	// fmt.Println(o.DecompressH.Stats())
	// fmt.Println("Write:", o.WriteH.Bins)
	// fmt.Println(o.WriteH.Stats())
	// fmt.Println(o.Wait1, o.Wait2, o.Wait3, o.Wait4)
	// fmt.Println()

	o.Mutex.Lock()
	Valid := o.Valid
	o.Mutex.Unlock()

	if Valid {
		<-o.WriteCompleteChan
		return o.BytesRcvdCompressed, o.BytesRcvdDecompressed, o.StoredError
	}
	return 0, 0, errors.New("Not active.")
}
