package walktree

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type WalkTreeCallbackInterface interface {
	PreDirectory(Path string) bool
	Directory(Path string)
	File(Path string, f os.FileInfo)
	DirectoryError(Path string, err error) bool
	NonFileObject(Path string, f os.FileInfo)
}

type WalkTreeStruct struct {
	Callback       WalkTreeCallbackInterface
	FollowSymlinks bool
}

func (o WalkTreeStruct) Sweep(Path string) bool {

	if !o.Callback.PreDirectory(Path) {
		return true
	}
	f, err := os.Open(Path)
	if err != nil {
		return o.Callback.DirectoryError(Path, err)
	}
	defer f.Close()

	o.Callback.Directory(Path)

	for {
		files, err := f.Readdir(1)

		if err != nil {
			if err == io.EOF {
				break
			}
			//			fmt.Println(files[0].Name())

			return o.Callback.DirectoryError(Path, err)

		}

		file := files[0]
		Fullname := filepath.Join(Path, file.Name())
		Mode := file.Mode()

		const RejectMask = fs.ModeCharDevice | fs.ModeDevice | fs.ModeNamedPipe | fs.ModeSocket

		if (Mode & RejectMask) != 0 {
			o.Callback.NonFileObject(Fullname, file)
			continue
		}

		if (Mode & fs.ModeSymlink) != 0 {
			if !o.FollowSymlinks {
				continue
			}

			fn, file, err := ParseSymlink(Fullname, file)
			if err != nil {
				o.Callback.DirectoryError(Fullname, err)
			}
			if fn == false {
				continue
			}

			Mode = file.Mode()
		}

		if (Mode & fs.ModeDir) != 0 {
			rc := o.Sweep(Fullname)
			if rc == false {
				return false
			}
		} else {
			o.Callback.File(Fullname, file)
		}
	}

	return true
}

func ParseSymlink(Filename string, f os.FileInfo) (bool, os.FileInfo, error) {
	Target, err := os.Readlink(Filename)
	if err != nil {
		return false, nil, err
	}

	if Target == "." {
		// Circular reference
		return false, nil, nil
	}

	Target = filepath.Clean(Target)

	f, err = os.Stat(Target)
	if err != nil {
		return false, nil, err
	}

	if f.Mode().IsDir() || f.Mode().IsRegular() {
		return true, f, nil
	}

	return false, nil, nil
}
