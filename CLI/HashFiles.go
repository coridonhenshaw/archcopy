package main

import (
	"Archcopy/HashFile"
	"ArchcopyCLI/util"
	"fmt"
)

func Blake2bHash() int {

	Input, err := SetupTransport(UI.RPC.InputURL)
	util.Fatal(err, "")
	defer Input.Close()

	d := func(Filename string, Filesize int64, Hash string) {
		fmt.Printf("%v %v\n", Hash, Filename)
	}

	for _, v := range UI.InputFile {
		result, err := HashFile.HashFile(v, nil)
		if err != nil {
			fmt.Println("Failed:", err)
		} else {
			d(v, 0, result)
		}
	}
	if len(UI.InputDirectory) > 0 {
		Input.FSAccess.SweepTree(UI.InputDirectory, true, true, d)
	}
	fmt.Println()
	return 0
}
