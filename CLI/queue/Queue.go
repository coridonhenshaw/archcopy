package queue

import "sort"

type QueueEntryStruct struct {
	InputFile      string
	InputFileSize  uint64
	OutputFile     string
	OutputFileSize uint64
	InputMissing   bool
	OutputExists   bool

	Invalid bool
	Failed  bool
}

type QueueStruct struct {
	List []QueueEntryStruct

	TotalInput int
	TotalValid int
	Rejected   int
	TotalBytes uint64
}

func (a QueueStruct) Len() int      { return len(a.List) }
func (a QueueStruct) Swap(i, j int) { a.List[i], a.List[j] = a.List[j], a.List[i] }
func (a QueueStruct) Less(i, j int) bool {
	return (a.List[i].Invalid != a.List[j].Invalid)
}

func (o *QueueStruct) DropInvalid(Test func(v QueueEntryStruct) bool) {
	o.TotalInput = len(o.List)

	for i, v := range o.List {
		o.List[i].Invalid = Test(v) // (v.InputMissing || v.OutputExists)
	}

	sort.Stable(o)
	i := sort.Search(len(o.List), func(i int) bool { return o.List[i].Invalid == true })
	o.List = o.List[0:i]

	o.TotalValid = len(o.List)
	o.Rejected = o.TotalInput - o.TotalValid

	for _, v := range o.List {
		o.TotalBytes += v.InputFileSize
	}
}

func (o *QueueStruct) Add(a QueueEntryStruct) {
	o.List = append(o.List, a)
}
