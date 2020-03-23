package mockclientlib

import "fmt"

func Logf(format string, args ...interface{}) {
	if log == nil {
		fmt.Printf(format+"\n", args...)
	} else {
		log.Infof(format, args...)
	}
}
