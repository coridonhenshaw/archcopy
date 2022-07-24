//go:build linux
// +build linux

package HashFile

import (
	"io"
	"os"
	"syscall"
	"time"

	"github.com/ncw/directio"
)

func Read(Filename string, Callback func(Buffer []byte, Progress ProgressStruct) error) error {

	var Progress ProgressStruct
	Progress.StartTime = time.Now()

	var err error
	var fh *os.File
	fh, err = directio.OpenFile(Filename, os.O_RDONLY, 0666)
	e, ok := err.(*os.PathError)
	if ok && e.Err == syscall.EINVAL { // No O_DIRECT support
		fh, err = os.OpenFile(Filename, os.O_RDONLY, 0666)
	}
	if err != nil {
		return err
	}
	defer fh.Close()

	Blocksize := 65536
	Buffer := directio.AlignedBlock(Blocksize)

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
		} else if err != nil {
			e, ok := err.(*os.PathError)
			if ok && e.Err == syscall.EINVAL { // Filesystem lied about O_DIRECT support.
				fh.Close()
				fh, err = os.OpenFile(Filename, os.O_RDONLY, 0666)
				if err != nil {
					return err
				}
				defer fh.Close()
				_, err = fh.Seek(Progress.TotalBytesRead, 0)
				if err != nil {
					return err
				}
				continue
			} else {
				return err
			}
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
