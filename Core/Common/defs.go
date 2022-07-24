package common

import (
	"io/fs"
	"os"
	"time"
)

type ExistAction int

const (
	Fail      ExistAction = 0
	Resume                = 1
	Overwrite             = 2
)

type FSAccess interface {
	CheckExists(Accessor func(i int) string, Setter func(i int, Exists bool, Size uint64)) error
	SweepTree(Path string, FollowSymlinks bool, GenerateHashes bool,
		Setter func(Filename string, Filesize int64, Hash string)) error
	RenameFile(Source string, Destination string) error
	GetFileHash(Source string) (string, error)

	GetFileWriter() FileWriterInterface
	GetFileReader() FileReaderInterface

	Open()
	Close()
}

type FilemetadataStruct struct {
	Permissions fs.FileMode
	Owner       *OwnershipStruct
	Atime       time.Time
	Mtime       time.Time
	Size        int64
}

type TransferProgressStruct struct {
	StartTime time.Time
	EndTime   time.Time

	TransferredBytes           uint64
	TransferredBytesCompressed uint64

	Final bool

	ReceivedHash string
	ReadbackHash string
}

type FileWriterInterface interface {
	CreateFile(DirectoryPerm os.FileMode, Fullname string, perm fs.FileMode, Owner *OwnershipStruct, Atime time.Time, Mtime time.Time, Size int64, Sparse bool,
		OnExist ExistAction, ReadbackHash bool, Progress chan TransferProgressStruct) error
	Write(Buffer *[]byte) error
	Finalize(Cancel bool) (TransferProgressStruct, error)
	UseReadHash() bool
}

type FileReaderInterface interface {
	Open(Filename string, ReadHash bool, Position int64, Progress chan TransferProgressStruct) (FilemetadataStruct, error)
	Read() ([]byte, error)
	Close() (TransferProgressStruct, error)
}

type OwnershipStruct struct {
	UID int
	GID int
}
