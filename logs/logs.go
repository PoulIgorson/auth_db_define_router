package logs

import (
	"fmt"
	"time"
)

func getNow() string {
	now := time.Now()
	return now.Format(time.DateTime)
}

const (
	NOTSET = 10 * iota
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)

var stringLevel = map[int]string{
	NOTSET:   "NOTSET",
	DEBUG:    "DEBUG",
	INFO:     "INFO",
	WARNING:  "WARNING",
	ERROR:    "ERROR",
	CRITICAL: "CRITICAL",
}

var Level = NOTSET

type log struct {
	datetime string
	level    int
	msg      string
}

func length(str string, l int) string {
	if len(str) >= l {
		return str
	}
	c := (l - len(str)) / 2
	n := ""
	for i := 0; i < c; i++ {
		n += " "
	}
	n += str
	for i := 0; i < c; i++ {
		n += " "
	}
	if len(n) < l {
		n += " "
	}
	return n
}

func (log *log) String() string {
	return fmt.Sprintf("%v | %v: %v", log.datetime, length(stringLevel[log.level], 10), log.msg)
}

func Log(level int, msg string, a ...any) {
	fmt.Println(log{getNow(), level, fmt.Sprintf(msg, a...)})
}

func Debug(msg string, a ...any) {
	if Level != DEBUG {
		return
	}
	fmt.Println(log{getNow(), DEBUG, fmt.Sprintf(msg, a...)})
}

func Info(msg string, a ...any) {
	if Level != INFO {
		return
	}
	fmt.Println(log{getNow(), INFO, fmt.Sprintf(msg, a...)})
}

func Warning(msg string, a ...any) {
	if Level != WARNING {
		return
	}
	fmt.Println(log{getNow(), WARNING, fmt.Sprintf(msg, a...)})
}

func Error(msg string, a ...any) {
	if Level != ERROR {
		return
	}
	fmt.Println(log{getNow(), ERROR, fmt.Sprintf(msg, a...)})
}

func Critical(msg string, a ...any) {
	if Level != CRITICAL {
		return
	}
	fmt.Println(log{getNow(), CRITICAL, fmt.Sprintf(msg, a...)})
}
