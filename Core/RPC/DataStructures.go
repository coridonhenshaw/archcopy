package rpc

import (
	common "Archcopy/Common"
	"encoding/json"
)

type FileMetadataStruct struct {
	Filename     string
	Permissions  uint64
	UID          int
	GID          int
	ATime        int64
	MTime        int64
	Size         int64
	Sparse       bool
	OnExist      common.ExistAction
	ReadbackHash bool
	Signature    []byte
}

func (o *FileMetadataStruct) Encode() (string, error) {
	Str, err := json.Marshal(*o)
	return string(Str), err
}

func (o *FileMetadataStruct) Decode(In string) error {
	return json.Unmarshal([]byte(In), o)
}
