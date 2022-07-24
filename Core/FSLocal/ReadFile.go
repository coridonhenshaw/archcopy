package FSLocal

import (
	common "Archcopy/Common"
	"io"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/djherbis/times"
)

const BufferSize = 65536 * 8

type ReadFileStruct struct {
	fh *os.File

	results   common.TransferProgressStruct
	starttime time.Time
	progress  chan common.TransferProgressStruct

	readhash   bool
	hashThread *common.HashThreadStruct
}

func (o *ReadFileStruct) Open(Filename string,
	ReadHash bool,
	Position int64,
	Progress chan common.TransferProgressStruct) (common.FilemetadataStruct, error) {
	var err error
	var Result common.FilemetadataStruct

	o.starttime = time.Now()
	o.progress = Progress
	o.readhash = ReadHash

	o.fh, err = os.Open(Filename)
	if err != nil {
		return Result, err
	}

	if Position != 0 {
		_, err = o.fh.Seek(Position, 0)
		if err != nil {
			return Result, err
		}
	}

	Stat, err := o.fh.Stat()
	if err != nil {
		return Result, err
	}

	Result.Size = Stat.Size()
	Result.Permissions = Stat.Mode().Perm()

	Times, err := times.Stat(Filename)
	if err != nil {
		return Result, err
	}
	Result.Atime = Times.AccessTime()
	Result.Mtime = Times.ModTime()

	file_sys := Stat.Sys()
	Stat_t, Valid := file_sys.(*syscall.Stat_t)
	if Valid {
		Result.Owner = &common.OwnershipStruct{GID: int(Stat_t.Gid), UID: int(Stat_t.Uid)}
	}

	if o.readhash {
		o.hashThread = common.GetHashThread(32)
		o.hashThread.Start()
	}

	return Result, err
}

func (o *ReadFileStruct) Read() ([]byte, error) {
	var err error
	Buffer := make([]byte, BufferSize)
	var BytesRead int
	BytesRead, err = o.fh.Read(Buffer)
	if BytesRead != BufferSize {
		Buffer = Buffer[:BytesRead]
	}
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}

	atomic.AddUint64(&o.results.TransferredBytes, uint64(BytesRead))
	atomic.AddUint64(&o.results.TransferredBytesCompressed, uint64(BytesRead))

	select {
	case o.progress <- o.results:
	default:
	}

	if o.readhash {
		o.hashThread.InputChan <- &Buffer
	}

	return Buffer, err
}

func (o *ReadFileStruct) Close() (common.TransferProgressStruct, error) {
	if o.results.Final {
		return o.results, nil
	}
	if o.readhash {
		o.results.ReceivedHash = o.hashThread.Await()
	}

	o.results.Final = true
	o.progress <- o.results
	return o.results, o.fh.Close()
}
