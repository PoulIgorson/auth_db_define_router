package log

import (
	"log"
	"os"
)

var LogError *log.Logger
var LogInfo *log.Logger

func init() {
	createLogInfo()
	createLogError()
}

func createLogError() {
	file, err := os.OpenFile("./error.log", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("log.createLogError:", err.Error())
		log.Println("log.createLogError: setted output to os.Stderr")
		file = os.Stderr
	}
	LogError = log.New(file, "ERROR", log.Lshortfile|log.LstdFlags|log.Lmsgprefix)
}

func createLogInfo() {
	file, err := os.OpenFile("./info.log", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("log.createLogInfo:", err.Error())
		log.Println("log.createLogInfo: setted output to os.Stderr")
		file = os.Stderr
	}
	LogInfo = log.New(file, "INFO", log.Lshortfile|log.LstdFlags|log.Lmsgprefix)
}
