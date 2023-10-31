package utils

import "log"

func ErrorRecover(message string) {

	if err := recover(); err != nil {
		log.Println("Recovered "+message+": ", err)
	}
}
