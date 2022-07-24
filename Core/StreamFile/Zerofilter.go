package streamfile

import "unsafe"

const SubBlockSize = 4096 * 4

//
// Returns true if input slice contains only zeros.
//
func SparseScan(Block *[SubBlockSize]byte) bool {
	var Accumulator uint64

	if len(*Block) != SubBlockSize {
		return false
	}

	const Limit = SubBlockSize / (int)(unsafe.Sizeof(Accumulator))

	Base := (*[Limit]uint64)(unsafe.Pointer(&(*Block)[0]))

	const Midpoint = Limit / 2
	const Final = Limit - 1

	if Base[0]|Base[Midpoint]|Base[Final] != 0 {
		return false
	}

	for i := 1; i < Limit; i = i + 1 {
		Accumulator = Accumulator | (*Base)[i]
	}

	return Accumulator == 0
}

func ZeroFilter(Buffer *[]byte, Span func(int, int, bool) error) error {
	var err error

	var sz = len(*Buffer)

	if sz < SubBlockSize {
		return Span(0, sz, false)
	}

	var HasTail = false
	if sz%SubBlockSize != 0 {
		HasTail = true
	}

	var limit = (sz / SubBlockSize) * SubBlockSize
	if limit == 1 {
		return Span(0, len(*Buffer), false)
	}

	var StartOffset = 0

	Base := (*[SubBlockSize]byte)(unsafe.Pointer(&(*Buffer)[0]))
	var BlockRunType = SparseScan(Base)

	var i int
	for i = SubBlockSize; i < limit; i += SubBlockSize {

		Base := (*[SubBlockSize]byte)(unsafe.Pointer(&(*Buffer)[i]))

		Sparse := SparseScan(Base)

		if Sparse != BlockRunType {
			err = Span(StartOffset, i, BlockRunType)
			if err != nil {
				return err
			}
			StartOffset = i
			BlockRunType = Sparse
		}
	}
	err = Span(StartOffset, i, BlockRunType)
	if err != nil {
		return err
	}
	if HasTail {
		err = Span(i, len(*Buffer), false)
	}

	return err
}
