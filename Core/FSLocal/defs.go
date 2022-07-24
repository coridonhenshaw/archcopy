package FSLocal

import (
	"os"
)

type LowerWriteDriver interface {
	SetFile(*os.File)
	Write(Buffer *[]byte) error
	Finalize() error
}
