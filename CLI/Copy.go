package main

import (
	Archcopy "Archcopy"
	common "Archcopy/Common"
	"ArchcopyCLI/queue"
	"ArchcopyCLI/util"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
)

func Copy() int {

	UI.HaveRoot = util.CheckForRoot()

	var err error

	Input, err := SetupTransport(UI.RPC.InputURL)
	util.Fatal(err, "")
	defer Input.Close()

	Output, err := SetupTransport(UI.RPC.OutputURL)
	util.Fatal(err, "")
	defer Output.Close()

	var Queue queue.QueueStruct
	var NoSourceCheck bool

	var BuildTargetPath func(Prefix string, InputFile string) string
	if UI.PreserveRelativePath {
		BuildTargetPath = func(Prefix string, InputFile string) string {
			return path.Join(UI.OutputDirectory, InputFile)
		}
	} else {

		BuildTargetPath = func(Prefix string, InputFile string) string {
			if len(Prefix) == 0 {
				Prefix, _ = path.Split(InputFile)
			}

			pth1 := strings.TrimPrefix(InputFile, Prefix)

			return path.Join(UI.OutputDirectory, pth1)
		}
	}

	if len(UI.InputFile) == 1 {
		var Target string

		if len(UI.OutputFile) == 0 && len(UI.OutputDirectory) > 0 {
			Target = BuildTargetPath("", UI.InputFile[0])
		} else if len(UI.OutputFile) > 0 && len(UI.OutputDirectory) == 0 {
			Target = UI.OutputFile
		} else {
			err = errors.New("Error: Incomprehensible combination of input and output parameters.")
		}
		Queue.Add(queue.QueueEntryStruct{InputFile: UI.InputFile[0], OutputFile: Target})
	} else if len(UI.InputFile) > 1 {
		if len(UI.OutputDirectory) == 0 {
			err = errors.New("Error: -od is required when multiple input files are specified.")
		} else {
			for _, v := range UI.InputFile {
				Queue.Add(queue.QueueEntryStruct{InputFile: v, OutputFile: path.Join(UI.OutputDirectory, v)})
			}
		}
	} else if len(UI.InputDirectory) > 0 {
		if len(UI.OutputDirectory) == 0 {
			err = errors.New("Error: -od is required with -id.")
		} else {
			Input.FSAccess.SweepTree(UI.InputDirectory, true, false,
				func(Filename string, Filesize int64, _ string) {
					QES := queue.QueueEntryStruct{InputFile: Filename,
						OutputFile:    BuildTargetPath(UI.InputDirectory, Filename),
						InputFileSize: uint64(Filesize)}
					Queue.Add(QES)
				})
			NoSourceCheck = true
		}
	} else {
		err = errors.New("Error: Incomprehensible combination of input and output parameters.")
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var ExistAction common.ExistAction
	if UI.AllowOverwrite {
		ExistAction = common.Overwrite
	} else if UI.Resume {
		ExistAction = common.Resume
	}

	if !NoSourceCheck {
		err := Input.FSAccess.CheckExists(
			func(i int) string {
				if i >= len(Queue.List) {
					return ""
				}
				return Queue.List[i].InputFile
			},
			func(i int, Exists bool, Size uint64) {
				Queue.List[i].InputMissing = !Exists
				Queue.List[i].InputFileSize = Size
			})

		if err != nil {
			util.Fatal(err, "Checking for input files")
		}
	}

	if ExistAction != common.Overwrite {
		err := Output.FSAccess.CheckExists(
			func(i int) string {
				if i >= len(Queue.List) {
					return ""
				}
				return Queue.List[i].OutputFile
			},
			func(i int, Exists bool, Size uint64) {
				if Exists {
					Queue.List[i].OutputExists = true
					Queue.List[i].OutputFileSize = Size
				}
			})

		if err != nil {
			util.Fatal(err, "Checking for output files")
		}
	}

	for _, v := range Queue.List {
		if v.InputMissing {
			fmt.Fprintf(os.Stderr, "Unable to access source %v\n", v.InputFile)
			if !UI.ContinueOnError {
				return 2
			}
		}
		if v.OutputExists && ExistAction == common.Fail {
			fmt.Fprintf(os.Stderr, "Pre-existing target %v\n", v.OutputFile)
			if !UI.ContinueOnError {
				return 2
			}
		}
	}

	var InvalidFunc func(v queue.QueueEntryStruct) bool

	switch ExistAction {
	case common.Fail:
		InvalidFunc = func(v queue.QueueEntryStruct) bool { return (v.InputMissing || v.OutputExists) }
	case common.Overwrite:
		InvalidFunc = func(v queue.QueueEntryStruct) bool { return (v.InputMissing) }
	case common.Resume:
		InvalidFunc = func(v queue.QueueEntryStruct) bool {
			return (v.InputMissing || (v.OutputExists && v.InputFileSize <= v.OutputFileSize))
		}
	}

	Queue.DropInvalid(InvalidFunc)

	if Queue.TotalValid <= 0 {
		fmt.Fprintln(os.Stderr, "Error: No files available to copy.")
		return 1
	} else if Queue.TotalValid > 0 {
		fmt.Printf("Transferring %v comprising %v.\n", english.Plural(Queue.TotalValid, "file", "files"), humanize.Bytes(uint64(Queue.TotalBytes)))
	}

	if Queue.Rejected > 0 {
		fmt.Printf(" Rejected %v.\n", english.Plural(Queue.Rejected, "file", "files"))
	}

	fmt.Println()

	if UI.DryRun {
		fmt.Print("Dry Run:\n\n")
		for _, v := range Queue.List {
			fmt.Printf("Copy %v %v => %v\n", humanize.Bytes(uint64(v.InputFileSize)), v.InputFile, v.OutputFile)
		}
		fmt.Printf("\n%v comprising %v would have been transferred.\n\n", english.Plural(Queue.TotalValid, "File", "Files"), humanize.Bytes(uint64(Queue.TotalBytes)))
		return 0
	}

	var Transferred int
	for i, v := range Queue.List {
		if len(v.InputFile) == 0 {
			continue
		}

		AC := Archcopy.CopyStruct{Source: v.InputFile, Target: v.OutputFile,
			ComputeReadbackHash: UI.Verify, Concurrent: true, Sparse: UI.Sparse,
			Control: CopyControlCallback, Progress: CopyProgressCallback, Chown: UI.HaveRoot,
			FSAccessRead: Input.FSAccess, FSAccessWrite: Output.FSAccess, ExistAction: ExistAction,
		}

		if ExistAction == common.Resume && v.OutputExists {
			AC.ResumeOffset = v.OutputFileSize
		}
		err := AC.CopyFile()

		if err == nil && UI.Verify {
			UIOutputFunc("Verified", AC.Source, AC.Target, AC.ExpectedBytes, AC.TransferredBytes,
				AC.Duration)
			Transferred++
		} else if err == nil && !UI.Verify {
			UIOutputFunc("Done", AC.Source, AC.Target, AC.ExpectedBytes, AC.TransferredBytes,
				AC.Duration)
			Transferred++
		} else {
			fmt.Fprintf(os.Stderr, "\r\x1b[2KFailed %v %v => %v\n", humanize.Bytes(uint64(AC.TotalBytes)), AC.Source, AC.Target)
			if err != nil {
				fmt.Fprintln(os.Stderr, "   ", err)
			}
			err = Output.FSAccess.RenameFile(AC.Target, AC.Target+".hashfailed")
			if err != nil {
				fmt.Fprintln(os.Stderr, "    Failed to rename:", err)
			}
			if !UI.ContinueOnError {
				fmt.Fprintln(os.Stderr, "Terminating")
				return 2
			}
			Queue.List[i].Failed = true
			continue
		}
	}

	if Transferred == Queue.TotalValid {
		if Transferred == 1 {
			fmt.Print("\nSuccess.\n\n")
		} else {
			fmt.Printf("\nSuccess: All %v expected files transferred.\n\n", Transferred)
		}
	} else {
		fmt.Printf("\nWarning: transferred only %v out of %v expected.\n\n", Transferred, Queue.TotalValid)
	}

	// if len(RejectQueue) > 0 {
	// 	fmt.Printf("Rejected: %v\n\n", len(RejectQueue))
	// 	for _, v := range RejectQueue {
	// 		fmt.Println("Rejected", v.InputFile)
	// 	}
	// 	fmt.Print("\n")
	// }

	// if len(Failed) > 0 {
	// 	fmt.Printf("Failed to transfer: %v\n\n", len(Failed))
	// 	for _, v := range Failed {
	// 		fmt.Println("Failed", v.InputFile)
	// 	}
	// 	os.Exit(1)
	// }

	return 0
}

func CopyControlCallback(o *Archcopy.CopyStruct, State string, err error) bool {
	fmt.Fprintln(os.Stderr, "\nError:", State, err)
	return false
}

func CopyProgressCallback(o *Archcopy.CopyStruct) bool {
	_, Source := path.Split(o.Source)
	Target, _ := path.Split(o.Target)
	Elapsed := time.Now().Sub(o.StartTime)

	UIOutputFunc("", Source, Target, o.ExpectedBytes, o.TransferredBytes, Elapsed)
	return true
}

func UIOutputFunc(Verb string, Source string, Target string,
	ExpectedBytes int, TransferredBytes int,
	ElapsedDuration time.Duration) {

	EB := humanize.Bytes(uint64(ExpectedBytes))
	TB := humanize.Bytes(uint64(TransferredBytes))
	El := ElapsedDuration.Truncate(1 * time.Second)
	Spd := humanize.Bytes(uint64(float64(TransferredBytes) / ElapsedDuration.Seconds()))

	if len(Verb) == 0 {
		fmt.Printf("%v / %v (%v %v/s) %v => %v\x1b[0K\r",
			TB,
			EB,
			El,
			Spd,
			Source,
			Target)
	} else {
		fmt.Printf("\r\x1b[2K%v %v (%v %v/s) %v => %v\n",
			Verb,
			EB,
			El,
			Spd,
			Source,
			Target)
	}
}
