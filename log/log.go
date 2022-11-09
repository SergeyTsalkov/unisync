package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	Debug = iota
	Notice
	Warn
	Fatal
)

type Output struct {
	Writer    io.Writer
	Level     uint8
	Timestamp string
}

var ScreenOutput io.Writer = os.Stdout
var ScreenLevel uint8 = Notice
var ScreenTimestmap = ""

var Outputs = []*Output{}

func writeTo(w io.Writer, ts, str string) error {
	var err error
	str = strings.TrimSpace(str)

	if ts != "" {
		ts = time.Now().Format(ts)
		_, err = io.WriteString(w, ts+" "+str+"\n")
	} else {
		_, err = io.WriteString(w, str+"\n")
	}
	return err
}

func write(level uint8, str string) error {
	if ScreenOutput != nil && level >= ScreenLevel {
		err := writeTo(ScreenOutput, ScreenTimestmap, str)
		if err != nil {
			return err
		}
	}

	for _, output := range Outputs {
		if level >= output.Level {
			err := writeTo(output.Writer, output.Timestamp, str)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Add(w io.Writer, level uint8, ts string) {
	Outputs = append(Outputs, &Output{w, level, ts})
}

func AddFile(fullpath string, level uint8, ts string) error {
	file, err := os.OpenFile(fullpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	Add(file, level, ts)
	return nil
}

func Reset() {
	Outputs = []*Output{}
}

// func GetLevel(w io.Writer) (uint8, bool) {
// 	for _, output := range Outputs {
// 		if output.Writer == w {
// 			return output.Level, true
// 		}
// 	}

// 	return 0, false
// }

func Debugln(a ...any) error {
	return write(Debug, fmt.Sprintln(a...))
}
func Println(a ...any) error {
	return write(Notice, fmt.Sprintln(a...))
}
func Warnln(a ...any) error {
	return write(Warn, fmt.Sprintln(a...))
}
func Fatalln(a ...any) {
	write(Fatal, fmt.Sprintln(a...))
	os.Exit(1)
}

func Debugf(format string, a ...any) error {
	return write(Debug, fmt.Sprintf(format, a...))
}
func Printf(format string, a ...any) error {
	return write(Notice, fmt.Sprintf(format, a...))
}
func Warnf(format string, a ...any) error {
	return write(Warn, fmt.Sprintf(format, a...))
}
func Fatalf(format string, a ...any) {
	write(Fatal, fmt.Sprintf(format, a...))
	os.Exit(1)
}
