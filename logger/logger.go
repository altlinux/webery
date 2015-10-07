/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package logger

import (
	log "github.com/Sirupsen/logrus"

	"fmt"
)

type ServerLogger struct {
	Subsys string
}

func (l *ServerLogger) Debug(msg string, args ...interface{}) {
	e := log.NewEntry(log.StandardLogger())

	for i := 0; i < len(args); i += 2 {
		k := fmt.Sprintf("%+v", args[i])
		e = e.WithField(k, args[i+1])
	}

	e.Debugf("[%s] %s", l.Subsys, msg)
}

func (l *ServerLogger) Info(msg string, args ...interface{}) {
	e := log.NewEntry(log.StandardLogger())

	for i := 0; i < len(args); i += 2 {
		k := fmt.Sprintf("%+v", args[i])
		e = e.WithField(k, args[i+1])
	}

	e.Infof("[%s] %s", l.Subsys, msg)
}

func (l *ServerLogger) Warn(msg string, args ...interface{}) {
	e := log.NewEntry(log.StandardLogger())

	for i := 0; i < len(args); i += 2 {
		k := fmt.Sprintf("%+v", args[i])
		e = e.WithField(k, args[i+1])
	}

	e.Warningf("[%s] %s", l.Subsys, msg)
}

func (l *ServerLogger) Error(msg string, args ...interface{}) {
	e := log.NewEntry(log.StandardLogger())

	for i := 0; i < len(args); i += 2 {
		k := fmt.Sprintf("%+v", args[i])
		e = e.WithField(k, args[i+1])
	}

	e.Errorf("[%s] %s", l.Subsys, msg)
}

func (l *ServerLogger) Fatal(msg string, args ...interface{}) {
	e := log.NewEntry(log.StandardLogger())

	for i := 0; i < len(args); i += 2 {
		k := fmt.Sprintf("%+v", args[i])
		e = e.WithField(k, args[i+1])
	}

	e.Fatalf("[%s] %s", l.Subsys, msg)
}
