package FSLocal

import (
	common "Archcopy/Common"
	"os"
)

type FSAccessLocalStruct struct {
}

func (o *FSAccessLocalStruct) CheckExists(Accessor func(i int) string, Setter func(i int, Exists bool, Size uint64)) error {

	for i := 0; ; i++ {
		Fn := Accessor(i)
		if len(Fn) == 0 {
			break
		}
		Stat, err := os.Stat(Fn)
		if err == nil {
			Setter(i, true, uint64(Stat.Size()))
		} else {
			Setter(i, false, 0)
		}

	}

	return nil
}

func (o *FSAccessLocalStruct) GetFileWriter() common.FileWriterInterface {
	return &LocalUpperWriteDriver{}
}

func (o *FSAccessLocalStruct) GetFileReader() common.FileReaderInterface {
	return &ReadFileStruct{}
}

func (o *FSAccessLocalStruct) RenameFile(Source string, Destination string) error {
	return os.Rename(Source, Destination)
}

func (o *FSAccessLocalStruct) Open() { return }

func (o *FSAccessLocalStruct) Close() { return }
