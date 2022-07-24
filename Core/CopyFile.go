package Archcopy

import (
	common "Archcopy/Common"
	"crypto/subtle"
	"errors"
	"time"
)

type CopyStruct struct {
	Source              string
	Target              string
	ComputeReadbackHash bool
	//	ComputeReadHash     bool
	Concurrent   bool
	Sparse       bool
	Chown        bool
	ExistAction  common.ExistAction
	ResumeOffset uint64

	FSAccessWrite common.FSAccess
	FSAccessRead  common.FSAccess
	//Callbacks
	Control  func(o *CopyStruct, State string, err error) bool
	Progress func(o *CopyStruct) bool
	//	Complete func(o *CopyStruct) bool
	//Results
	HashSource string
	//	HashReceived string
	HashReadback string

	StartTime                  time.Time
	Duration                   time.Duration
	ExpectedBytes              int
	TotalBytes                 int
	TransferredBytes           int
	TransferredBytesCompressed int
}

func (o *CopyStruct) CopyFile() error {
	if o.Control == nil {
		return errors.New("No control function provided.")
	}
	if o.Progress == nil {
		return errors.New("No progress function provided.")
	}
	if o.FSAccessWrite == nil {
		return errors.New("No FSAccess write object provided.")
	}
	if o.FSAccessRead == nil {
		return errors.New("No FSAccess read object provided.")
	}
	o.StartTime = time.Now()

	FileWriter := o.FSAccessWrite.GetFileWriter()
	FileReader := o.FSAccessRead.GetFileReader()

	// HashChannel := make(chan *[]byte, 32)
	// ResultChannel := make(chan string)

	// if o.ComputeReadHash {
	// 	go ComputeHashThread(HashChannel, ResultChannel)
	// }

	WriteChannel := make(chan *[]byte, 32)
	CompleteChannel := make(chan error)

	var Terminate bool
	WriteProgressChannel := make(chan common.TransferProgressStruct)
	ReadProgressChannel := make(chan common.TransferProgressStruct)
	go func() {
		var Done = 0
		var NextWindow = time.Now()
		for Done < 2 {
			select {
			case q := <-WriteProgressChannel:
				o.TransferredBytes = int(q.TransferredBytes)
				o.TransferredBytesCompressed = int(q.TransferredBytesCompressed)
				if time.Now().After(NextWindow) {
					NextWindow = time.Now().Add(125 * time.Millisecond)
					if !o.Progress(o) {
						Terminate = true
					}
				}
				if q.Final {
					Done++
				}
			case q := <-ReadProgressChannel:
				if q.Final {
					Done++
				}
			}
		}
	}()

	var SeekOffset int64
	if o.ResumeOffset != 0 && o.ExistAction == common.Resume {
		SeekOffset = int64(o.ResumeOffset)
	}

	var GetReadHash = o.ComputeReadbackHash || FileWriter.UseReadHash()

	Metadata, err := FileReader.Open(o.Source, GetReadHash, SeekOffset, ReadProgressChannel)

	if err != nil {
		o.Control(o, "Source access error", err)
		return err
	}
	defer FileReader.Close()

	o.ExpectedBytes = int(Metadata.Size)

	if !o.Chown {
		Metadata.Owner = nil
	}

	err = FileWriter.CreateFile(0700, o.Target, Metadata.Permissions,
		Metadata.Owner,
		Metadata.Atime,
		Metadata.Mtime,
		Metadata.Size,
		o.Sparse,
		o.ExistAction,
		o.ComputeReadbackHash,
		WriteProgressChannel)

	if err != nil {
		o.Control(o, "Target access error", err)
		return err
	}

	if o.Concurrent {
		go WriteThread(FileWriter, WriteChannel, CompleteChannel)
	}

	for Terminate == false {
		var Buffer []byte
		Buffer, err = FileReader.Read()

		var BytesRead int = len(Buffer)

		if BytesRead == 0 { // EOF
			Terminate = true
		} else if err != nil {
			o.Control(o, "Source access error (read)", err)
			Terminate = true
			continue
		}
		o.TotalBytes += BytesRead

		// if o.ComputeReadHash {
		// 	HashChannel <- &Buffer
		// }

		if o.Concurrent {
			select {
			case WriteChannel <- &Buffer:
			case err = <-CompleteChannel:
				Terminate = true
				continue
			}

		} else {
			err = FileWriter.Write(&Buffer)
			if err != nil {
				Terminate = true
				continue
			}
		}
	}

	if err != nil {
		o.Control(o, "Target access error", err)
		FileWriter.Finalize(true)
		return err
	}

	if o.Concurrent {
		WriteChannel <- nil
		err = <-CompleteChannel
		if err != nil {
			o.Control(o, "Target access error", err)
			FileWriter.Finalize(true)
			return err
		}
	}

	ReadResults, err1 := FileReader.Close()
	o.HashSource = ReadResults.ReceivedHash

	var Results common.TransferProgressStruct
	Results, err = FileWriter.Finalize(false)

	o.HashReadback = Results.ReadbackHash
	o.TransferredBytes = int(Results.TransferredBytes)
	o.TransferredBytesCompressed = int(Results.TransferredBytesCompressed)
	o.Duration = time.Now().Sub(o.StartTime)

	if err1 != nil {
		o.Control(o, "Source access error", err1)
		return err
	}

	if err != nil {
		o.Control(o, "Target access error", err)
		return err
	}

	if FileWriter.UseReadHash() {
		if subtle.ConstantTimeCompare([]byte(o.HashSource), []byte(Results.ReceivedHash)) == 0 {
			err = errors.New("data corrupted in transit.")
			o.Control(o, "Target access error", err)
			return err
		}
	}

	if o.ComputeReadbackHash {
		if subtle.ConstantTimeCompare([]byte(o.HashSource), []byte(Results.ReadbackHash)) == 0 {
			err = errors.New("readback failed")
			o.Control(o, "Target access error", err)
			return err
		}
	}

	return nil
}

func WriteThread(FileWriter common.FileWriterInterface, BufferChannel chan *[]byte, CompleteChannel chan error) {

	var err error

	for err == nil {
		Buffer := <-BufferChannel

		if Buffer == nil {
			break
		}

		err = FileWriter.Write(Buffer)
	}

	CompleteChannel <- err
}
