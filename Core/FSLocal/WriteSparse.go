package FSLocal

import (
	"log"
	"os"
	"unsafe"
)

const ClusterSize = 4096

type WriteSparseStruct struct {
	fh       *os.File
	IsSparse bool
}

func (o *WriteSparseStruct) SetFile(fh *os.File) {
	o.fh = fh
}

func SparseScanUnsafe(Block *[ClusterSize]byte) bool {

	if len(*Block) != ClusterSize {
		return false
	}

	var Accumulator uint64

	const Limit = ClusterSize / (int)(unsafe.Sizeof(Accumulator))

	Base := (*[Limit]uint64)(unsafe.Pointer(&(*Block)[0]))

	const Midpoint = Limit / 2
	const Final = Limit - 1

	if Base[0]+Base[Midpoint]+Base[Final] != 0 {
		return false
	}

	for i := 0; i < Limit; i = i + 1 {
		Accumulator = Accumulator | (*Base)[i]
	}

	return Accumulator == 0
}

func (o *WriteSparseStruct) Write(Buffer *[]byte) error {

	if Buffer == nil {
		log.Panic("Nil buffer")
	}

	var i int
	var sz = len(*Buffer)
	var limit = (sz / ClusterSize) * ClusterSize

	for i = 0; i < limit; i += ClusterSize {

		Base := (*[ClusterSize]byte)(unsafe.Pointer(&(*Buffer)[i]))

		o.IsSparse = SparseScanUnsafe(Base)

		if o.IsSparse == true {
			o.fh.Seek(ClusterSize, 1)
		} else {
			o.fh.Write((*Buffer)[i : i+ClusterSize])
		}
	}
	if i != sz {
		_, err := o.fh.Write((*Buffer)[i:])
		if err != nil {
			return err
		}
		o.IsSparse = false
	}

	return nil
}

func (o *WriteSparseStruct) Finalize() error {
	if o.IsSparse == true {
		pos, err := o.fh.Seek(0, 1)
		if err != nil {
			return err
		}
		pos, err = o.fh.Seek(pos-1, 0)
		if err != nil {
			return err
		}
		_, err = o.fh.Write(make([]byte, 1))
		if err != nil {
			return err
		}
	}
	return nil
}
