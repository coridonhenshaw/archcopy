package FSLocal

import (
	"os"
)

type WriteStandardStruct struct {
	fh *os.File
}

func (o *WriteStandardStruct) SetFile(fh *os.File) {
	o.fh = fh
}

func (o *WriteStandardStruct) Write(Buffer *[]byte) error {
	_, err := o.fh.Write(*Buffer)
	return err
}

func (o *WriteStandardStruct) Finalize() error {
	return nil
}
