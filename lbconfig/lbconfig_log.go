package lbconfig

import (
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"sync"
	"time"
)

//Log struct for the log
type Log struct {
	Writer     syslog.Writer
	Syslog     bool
	Stdout     bool
	Debugflag  bool
	TofilePath string
	logMu      sync.Mutex
}

//Logger struct for the Logger interface
type Logger interface {
	Info(s string) error
	Warning(s string) error
	Debug(s string) error
	Error(s string) error
}

//Write_to_log put something in the log file
func (lbc *LBConfig) Write_to_report(level string, msg string) error {

	myMessage := msg

	if level == "INFO" {
		lbc.Rlog.Info(myMessage)
	} else if level == "DEBUG" {
		lbc.Rlog.Debug(myMessage)
	} else if level == "WARN" {
		lbc.Rlog.Warning(myMessage)
	} else if level == "ERROR" {
		lbc.Rlog.Error(myMessage)
	} else {
		lbc.Rlog.Error("LEVEL " + level + " NOT UNDERSTOOD, ASSUMING ERROR " + myMessage)
	}

	return nil
}

//Info write as Info
func (l *Log) Info(s string) error {
	var err error
	if l.Syslog {
		err = l.Writer.Info(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("INFO " + s)
	}
	return err

}

//Warning write as Warning
func (l *Log) Warning(s string) error {
	var err error
	if l.Syslog {
		err = l.Writer.Warning(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("WARN " + s)
	}
	return err

}

//Debug write as Debug
func (l *Log) Debug(s string) error {
	var err error
	if l.Debugflag {
		if l.Syslog {
			err = l.Writer.Debug(s)
		}
		if l.Stdout || (l.TofilePath != "") {
			err = l.writefilestd("DEBUG " + s)
		}
	}
	return err

}

//Error write as Error
func (l *Log) Error(s string) error {
	var err error
	if l.Syslog {
		err = l.Writer.Err(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("ERROR " + s)
	}
	return err

}

func (l *Log) writefilestd(s string) error {
	var err error
	nl := ""
	if !strings.HasSuffix(s, "\n") {
		nl = "\n"
	}
	t := time.Now()
	timestamp := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
		t.Day(), t.Month(), t.Year(),
		t.Hour(), t.Minute(), t.Second())
	msg := fmt.Sprintf("%s %s%s",
		timestamp,
		s, nl)
	l.logMu.Lock()
	defer l.logMu.Unlock()
	if l.Stdout {
		_, err = fmt.Printf(msg)
	}
	if l.TofilePath != "" {
		f, err := os.OpenFile(l.TofilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = fmt.Fprintf(f, msg)
	}
	return err
}
