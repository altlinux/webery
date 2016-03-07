package logger

import (
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func GetFileLine(depth int) (file string, line int, ok bool) {
	_, file, line, ok = runtime.Caller(depth)

	if ok {
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}
	} else {
		file = "???"
		line = 0
	}
	return
}

func WithFieldsDepth(args log.Fields, depth int) (e *log.Entry) {
	e = log.NewEntry(log.StandardLogger())
	e = e.WithFields(args)

	if file, line, ok := GetFileLine(depth); ok {
		e = e.WithField("file", file)
		e = e.WithField("fileline", line)
	}
	return
}

func WithFields(args log.Fields) (e *log.Entry) {
	e = WithFieldsDepth(args, 3)
	return
}
