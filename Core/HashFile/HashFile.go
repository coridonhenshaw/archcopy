package HashFile

import (
	common "Archcopy/Common"
	"time"
)

type ProgressStruct struct {
	StartTime      time.Time
	TotalBytesRead int64
	TotalFileSize  int64
}

func HashFile(Filename string, ProgressChan chan *ProgressStruct) (string, error) {

	HThread := common.GetHashThread(32)
	defer HThread.Cancel()
	HThread.Start()

	Callback := func(Buffer []byte, Progress ProgressStruct) error {
		HThread.Write(&Buffer)
		if ProgressChan != nil {
			select {
			case ProgressChan <- &Progress:
			default:
			}
		}
		return nil
	}

	err := Read(Filename, Callback)

	if ProgressChan != nil {
		ProgressChan <- nil
	}

	Result := HThread.Await()

	return Result, err
}
