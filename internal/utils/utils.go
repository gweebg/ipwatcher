package utils

import "log"

func Check(err error, format string, args ...interface{}) {
	if err != nil {
		if format != "" {
			log.Fatalf(format, args...)
		} else {
			log.Fatal(err.Error())
		}
	}
}
