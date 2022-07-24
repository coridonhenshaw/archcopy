package FSLocal

import (
	Common "Archcopy/Common"
	common "Archcopy/Common"
	"Archcopy/HashFile"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type LocalUpperWriteDriver struct {
	Filename     string
	fh           *os.File
	wr           LowerWriteDriver
	Atime        time.Time
	Mtime        time.Time
	ReadbackHash bool

	result   common.TransferProgressStruct
	progress chan common.TransferProgressStruct
}

func (o *LocalUpperWriteDriver) UseReadHash() bool { return false }

func (o *LocalUpperWriteDriver) CreateFile(DirectoryPerm os.FileMode, Fullname string, perm fs.FileMode,
	Owner *common.OwnershipStruct,
	Atime time.Time, Mtime time.Time, Size int64,
	Sparse bool, OnExist common.ExistAction, ReadbackHash bool,
	Progress chan common.TransferProgressStruct) error {
	var err error

	o.Atime = Atime
	o.Mtime = Mtime
	o.Filename = Fullname
	o.ReadbackHash = ReadbackHash
	o.result.StartTime = time.Now()
	o.progress = Progress

	if Sparse {
		o.wr = new(WriteSparseStruct)
	} else {
		o.wr = new(WriteStandardStruct)
	}

	var flag int
	switch OnExist {
	case Common.Fail:
		flag = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	case Common.Resume:
		flag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	case Common.Overwrite:
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}

	stat, err := os.Stat(Fullname)
	if err == nil { //File exists
		switch OnExist {
		case Common.Fail:
			return errors.New("File exists.")
		case Common.Resume:
			if stat.Size() >= Size {
				return errors.New("Resume target larger or equal to source file.")
			}
		}
	}

	Directory := filepath.Dir(Fullname)
	err = os.MkdirAll(Directory, DirectoryPerm)
	if err != nil {
		return err
	}

	o.fh, err = os.OpenFile(Fullname, flag, perm)
	if err != nil {
		return err
	}

	if Owner != nil {
		err = o.fh.Chown(Owner.UID, Owner.GID)
		if err != nil {
			return err
		}
	}

	o.wr.SetFile(o.fh)

	return err
}

func (o *LocalUpperWriteDriver) Write(Buffer *[]byte) error {
	atomic.AddUint64(&o.result.TransferredBytes, uint64(len(*Buffer)))
	atomic.AddUint64(&o.result.TransferredBytesCompressed, uint64(len(*Buffer)))
	select {
	case o.progress <- o.result:
	default:
	}

	return o.wr.Write(Buffer)
}

func (o *LocalUpperWriteDriver) SetPosition(Position int64) error {
	_, err := o.fh.Seek(Position, 0)
	return err
}

func (o *LocalUpperWriteDriver) Finalize(Cancel bool) (common.TransferProgressStruct, error) {
	o.wr.Finalize()
	o.fh.Sync()
	o.fh.Close()

	var err error
	if o.ReadbackHash == true && Cancel == false {
		o.result.ReadbackHash, err = HashFile.HashFile(o.Filename, nil)
		if err != nil {
			return o.result, err
		}
	}

	o.result.Final = true
	o.progress <- o.result

	os.Chtimes(o.Filename, o.Atime, o.Mtime)
	if err != nil {
		return o.result, err
	}

	return o.result, err
}
