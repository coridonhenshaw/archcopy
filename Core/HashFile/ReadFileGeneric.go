// go:build !linux
//go:build !linux
// +build !linux

package HashFile

import (
	"io"
	"os"
	"time"
)

func Read(Filename string, Callback func(Buffer []byte, Progress ProgressStruct) error) error {

	var Progress ProgressStruct
	Progress.StartTime = time.Now()

	var err error
	var fh *os.File
	fh, err = os.OpenFile(Filename, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer fh.Close()

	Buffer := make([]byte, 65536)

	fs, err := fh.Stat()
	if err != nil {
		return err
	}

	Progress.TotalFileSize = fs.Size()

	var Terminate bool

	for Terminate == false {

		BytesRead, err := io.ReadFull(fh, Buffer)

		if err == io.ErrUnexpectedEOF || err == io.EOF {
			Buffer = Buffer[:BytesRead]
			Terminate = true
		}

		OutBuf := make([]byte, len(Buffer))
		copy(OutBuf, Buffer)

		err = Callback(OutBuf, Progress)
		if err != nil {
			return err
		}
		Progress.TotalBytesRead += int64(BytesRead)
	}
	err = Callback(nil, Progress)
	return err
}
