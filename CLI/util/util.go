package util

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
)

func CheckForRoot() bool {
	if runtime.GOOS == "linux" {
		u, _ := user.Current()
		return u.Uid == "0"
	}
	return false
}

var StackTrace = true

func Fatal(err error, Circumstance string) {
	if err != nil {
		if StackTrace {
			log.Panicf("%s: %s", Circumstance, err)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s", Circumstance, err)
		}
	}
}

func Base64(In []byte) string {
	var RawURLEncoding = base64.URLEncoding.WithPadding(base64.NoPadding)
	return RawURLEncoding.EncodeToString([]byte(In))
}
