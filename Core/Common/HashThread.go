package common

import (
	"encoding/hex"
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/blake2b"
)

type HashThreadStruct struct {
	InputChan  chan (*[]byte)
	ResultChan chan (string)
	CancelChan chan (bool)
	valid      bool
}

func GetHashThread(Depth int) *HashThreadStruct {
	var o HashThreadStruct

	o.InputChan = make(chan *[]byte, Depth)
	o.ResultChan = make(chan string)
	o.CancelChan = make(chan bool)

	return &o
}

func (o *HashThreadStruct) Start() {
	o.valid = true
	go o.Thread()
}

func (o *HashThreadStruct) Write(Block *[]byte) error {
	if !o.valid {
		return errors.New("Not active")
	}
	o.InputChan <- Block
	return nil
}

func (o *HashThreadStruct) Await() string {
	if o.valid == false {
		return ""
	}
	o.valid = false

	select {
	case o.InputChan <- nil:
	default:
	}

	s := <-o.ResultChan

	return s
}

func (o *HashThreadStruct) Cancel() {
	if o.valid {
		o.CancelChan <- true
	}
}

func (o *HashThreadStruct) Thread() {
	h, err := blake2b.New(32, nil)
	if err != nil {
		log.Panic(err)
	}

Outer:
	for {
		Timeout := time.NewTimer(2 * time.Minute)
		TimeoutChan := Timeout.C

		var Buffer *[]byte

		select {
		case <-o.CancelChan:
			break Outer
		case Buffer = <-o.InputChan:
			if Buffer == nil || len(*Buffer) == 0 {
				break Outer
			}
			h.Write(*Buffer)
		case <-TimeoutChan:
			break Outer
		}
		Timeout.Stop()
	}

	o.ResultChan <- hex.EncodeToString(h.Sum(nil))
}
