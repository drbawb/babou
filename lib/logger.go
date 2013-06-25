package lib

import (
	"log"
)

var DEBUG_MSG_ON bool = true

func Printf(format string, args ...interface{}) {
	if DEBUG_MSG_ON {
		log.Printf("[DEBUG] "+format, args)
	}
}

func Println(messages ...interface{}) {
	if DEBUG_MSG_ON {
		log.Println(messages...)
	}
}

func Print(messages ...interface{}) {
	if DEBUG_MSG_ON {
		log.Print(messages...)
	}
}
