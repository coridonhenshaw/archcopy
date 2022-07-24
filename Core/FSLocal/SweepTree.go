package FSLocal

import (
	walktree "Archcopy/FSLocal/WalkTree"
	"Archcopy/HashFile"
	"errors"
	"io/fs"
	"os"
)

type WalkTreeCallbackStruct struct {
	Setter         func(Filename string, Filesize int64, Hash string)
	GenerateHashes bool
}

func (o *FSAccessLocalStruct) SweepTree(Path string, FollowSymlinks bool, GenerateHashes bool,
	Setter func(Filename string, Filesize int64, hash string)) error {

	WalkTreeCallback := WalkTreeCallbackStruct{Setter: Setter, GenerateHashes: GenerateHashes}
	WalkTree := walktree.WalkTreeStruct{Callback: &WalkTreeCallback, FollowSymlinks: FollowSymlinks}

	WalkTree.Sweep(Path)

	return nil
}

func (o *WalkTreeCallbackStruct) NonFileObject(p string, f os.FileInfo) {
	return
}

func (o *WalkTreeCallbackStruct) File(p string, f os.FileInfo) {
	var Hash string
	//	var err error
	if o.GenerateHashes {
		Hash, _ = HashFile.HashFile(p, nil)
	}
	o.Setter(p, f.Size(), Hash)
}

func (o *WalkTreeCallbackStruct) PreDirectory(p string) bool {
	// return false to halt sweep
	return true
}

func (o *WalkTreeCallbackStruct) Directory(p string) {
	return
}

func (o *WalkTreeCallbackStruct) DirectoryError(Fullname string, err error) bool {
	// return false to halt sweep
	if errors.Is(err, fs.ErrPermission) {
		return true
	}

	if errors.Is(err, fs.ErrNotExist) {
		return true
	}

	return false
}
