package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/jacktrane/gocomponent/file_util"
	"github.com/jacktrane/gocomponent/time_format"
)

const (
	PanicLevel int = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

// 等级最高 则数值最低
var printLevelArr = map[int]string{
	0: "[Panic] ",
	1: "[Fatal] ",
	2: "[Error] ",
	3: "[Warn] ",
	4: "[Info] ",
	5: "[Debug] ",
}

type LogFile struct {
	level    int
	logTime  int64
	fileName string
	fileFd   *os.File
}

var gLogFile LogFile

func init() {
	NewConfig("", 5)
}

func NewConfig(logFolder string, level int) {
	gLogFile.fileName = logFolder
	gLogFile.level = level
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	if logFolder != "" {
		log.SetOutput(io.MultiWriter(os.Stderr, gLogFile))
		createFile(gLogFile.fileName, gLogFile.fileFd) // 在初始化时先加个fd先
	}
}

func SetLevel(level int) {
	gLogFile.level = level
}

func Debugf(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Debug] ")
		log.Output(2, fmt.Sprintf(format, args...))
	}
}

func Debug(v ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Debug] ")
		log.Output(2, fmt.Sprint(v...))
	}
}

func Infof(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Info] ")
		log.Output(2, fmt.Sprintf(format, args...))
	}
}

func Info(v ...interface{}) {
	if gLogFile.level >= InfoLevel {
		log.SetPrefix("[Info] ")
		log.Output(2, fmt.Sprint(v...))
	}
}

func Warnf(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Warn] ")
		log.Output(2, fmt.Sprintf(format, args...))
	}
}

func Warn(v ...interface{}) {
	if gLogFile.level >= WarnLevel {
		log.SetPrefix("[Warn] ")
		log.Output(2, fmt.Sprint(v...))
	}
}

func Errorf(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Error] ")
		log.Output(2, fmt.Sprintf(format, args...))
	}
}

func Error(v ...interface{}) {
	if gLogFile.level >= ErrorLevel {
		log.SetPrefix("[Error] ")
		log.Output(2, fmt.Sprint(v...))
	}
}

func Fatalf(format string, args ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Fatal] ")
		log.Output(2, fmt.Sprintf(format, args...))
		debug.PrintStack()
		os.Exit(1)
	}
}

func Fatal(v ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Fatal] ")
		log.Output(2, fmt.Sprint(v...))
		debug.PrintStack()
		os.Exit(1)
	}
}

func Panicf(format string, args ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Panic] ")
		log.Panic(fmt.Sprintf(format, args...))
	}
}

func Panic(v ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Panic] ")
		log.Panic(fmt.Sprint(v...))
	}
}

func (me LogFile) Write(buf []byte) (n int, err error) {
	if me.fileName == "" {
		fmt.Printf("consol: %s", buf)
		return len(buf), nil
	}

	if gLogFile.logTime+3600 < time_format.GetNowTime().Unix() {
		gLogFile.createLogFile()
		gLogFile.logTime = time_format.GetNowTime().Unix()
	}

	if gLogFile.fileFd == nil {
		return len(buf), nil
	}

	return gLogFile.fileFd.Write(buf)
}

func (me *LogFile) createLogFile() {
	if index := strings.LastIndex(me.fileName, "/"); index != -1 {
		os.MkdirAll(me.fileName[0:index], os.ModePerm)
	}

	now := time_format.GetNowTime()
	err, fileModTime := file_util.GetFileModTime(me.fileName)
	if err != nil {
		fmt.Println(err, me.fileName)
	}

	if err != nil || now.Format(time_format.FullFormatDateSimpleDay) != fileModTime.Format(time_format.FullFormatDateSimpleDay) {
		fmt.Println("log")
		d, _ := time.ParseDuration("-24h")
		beforeDay := now.Add(d)
		filename := fmt.Sprintf("%s_%s.log", me.fileName, beforeDay.Format(time_format.FullFormatDateSimpleDay))
		if !file_util.IsExist(filename) {
			if err := os.Rename(me.fileName, filename); err == nil {
				// go func() {
				// 	tarCmd := exec.Command("tar", "-zcf", filename+".tar.gz", filename, "--remove-files")
				// 	tarCmd.Run()

				// 	rmCmd := exec.Command("/bin/sh", "-c", "find "+logdir+` -type f -mtime +2 -exec rm {} \;`)
				// 	rmCmd.Run()
				// }()
			}
		}
	}

	createFile(me.fileName, me.fileFd)
}

func createFile(fileName string, fileFd *os.File) {
	for index := 0; index < 10; index++ {
		if fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm); nil == err {
			fileFd.Sync()
			fileFd.Close()
			fileFd = fd

			// 下面是为了重定向标准输出到文件中，因为painc，Dup2仅能在linux运行哦，所以如果在window下注释
			syscall.Dup2(int(fileFd.Fd()), int(os.Stdout.Fd()))
			syscall.Dup2(int(fileFd.Fd()), int(os.Stderr.Fd()))
			break
		}

		fileFd = nil
	}
}
