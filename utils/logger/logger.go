package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

type LogLevel int

var levelFlags = []string{"", "Fatal", "Error", "Warn", "Info", "Debug"}

const (
	maxLogNum   = 1e5
	callerDepth = 2

	RESET  = "\033[0m"
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"

	LOG_FILE_FORMAT = "%s-%s.%s"
)

const (
	NULL LogLevel = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG
)

type Settings struct {
	Path  string `yaml:"path"`
	Name  string `yaml:"name"`
	Ext   string `yaml:"ext"`
	DateF string `yaml:"date_format"`
}

type LogMessage struct {
	level LogLevel
	msg   string
}

func (l *LogMessage) reset() {
	l.level = NULL
	l.msg = ""
}

type logger struct {
	logFile           *os.File
	logStd            *log.Logger
	logMessageChannel chan *LogMessage
	logMessagePoll    *sync.Pool
	logLevel          LogLevel
	close             chan struct{}
}

func (l *logger) Close() {
	close(l.close)
}

func (l *logger) writeLog(level LogLevel, callerDepth int, msg string) {
	var formattedMsg string
	_, file, line, ok := runtime.Caller(callerDepth)
	if ok {
		formattedMsg = fmt.Sprintf("[%s][%s:%d] %s", levelFlags[level], file, line, msg)
	} else {
		formattedMsg = fmt.Sprintf("[%s] %s", levelFlags[level], msg)
	}
	logMsg := l.logMessagePoll.Get().(*LogMessage)
	logMsg.level = level
	logMsg.msg = formattedMsg
	l.logMessageChannel <- logMsg
}

func Init() {
	Setup(&Settings{
		Path:  "./logs",
		Name:  "redis",
		Ext:   "log",
		DateF: time.Now().Format("200601021504"),
	})
	SetLoggerLevel(DEBUG)
}

var defaultLogger *logger = newDefaultLogger()

func newDefaultLogger() *logger {
	stdLogger := &logger{
		logFile:           nil,
		logStd:            log.New(os.Stdout, "", log.LstdFlags),
		logMessageChannel: make(chan *LogMessage, maxLogNum),
		logLevel:          DEBUG,
		close:             make(chan struct{}),
		logMessagePoll: &sync.Pool{
			New: func() interface{} {
				return &LogMessage{}
			},
		},
	}
	go func() {
		for {
			select {
			case <-stdLogger.close:
				return
			case logMsg := <-stdLogger.logMessageChannel:
				msg := logMsg.msg
				switch logMsg.level {
				case ERROR:
					msg = RED + msg + RESET
				case WARN:
					msg = YELLOW + msg + RESET
				case INFO:
					msg = GREEN + msg + RESET
				case DEBUG:
					msg = BLUE + msg + RESET
				}
				stdLogger.logStd.Output(0, msg)
				logMsg.reset()
				//复用
				stdLogger.logMessagePoll.Put(logMsg)
			}
		}
	}()
	return stdLogger
}

func newFileLogger(settings *Settings) (*logger, error) {
	fileName := fmt.Sprintf(LOG_FILE_FORMAT, settings.Name, time.Now().Format(settings.DateF), settings.Ext)
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("newFileLogger.OenFile err:%s", err)
	}

	fileLogger := &logger{
		logFile:           f,
		logStd:            log.New(os.Stdout, "", log.LstdFlags),
		logMessageChannel: make(chan *LogMessage, maxLogNum),
		logLevel:          DEBUG,
		close:             make(chan struct{}),
		logMessagePoll: &sync.Pool{
			New: func() any {
				return &LogMessage{}
			},
		},
	}

	go func() {
		for {
			select {
			case <-fileLogger.close:
				return
			case logMsg := <-fileLogger.logMessageChannel:
				logFileName := fmt.Sprintf(LOG_FILE_FORMAT, settings.Name, time.Now().Format(settings.DateF), settings.Ext)
				if path.Join(settings.Path, logFileName) != fileLogger.logFile.Name() {
					fd, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY, 0666)
					if err != nil {
						panic("open log " + fileName + "failed: " + err.Error())
					}
					fileLogger.logFile.Close()
					fileLogger.logFile = fd
				}
				msg := logMsg.msg
				switch logMsg.level {
				case ERROR:
					msg = RED + msg + RESET
				case WARN:
					msg = YELLOW + msg + RESET
				case INFO:
					msg = GREEN + msg + RESET
				case DEBUG:
					msg = BLUE + msg + RESET
				}

				fileLogger.logStd.Output(0, msg)
				fileLogger.logFile.WriteString(time.Now().Format(time.RFC3339) + " " + logMsg.msg + "\r\n")
			}
		}
	}()

	return fileLogger, nil
}

func Setup(s *Settings) {
	defaultLogger.Close()
	logger, err := newFileLogger(s)
	if err != nil {
		panic(err)
	}
	defaultLogger = logger
}

func SetLoggerLevel(logLevel LogLevel) {
	defaultLogger.logLevel = logLevel
}

func Debug(v ...any) {
	if defaultLogger.logLevel >= DEBUG {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(DEBUG, callerDepth, msg)
	}
}
func Debugf(format string, v ...any) {
	if defaultLogger.logLevel >= DEBUG {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(DEBUG, callerDepth, msg)
	}
}

func Info(v ...any) {
	if defaultLogger.logLevel >= INFO {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(INFO, callerDepth, msg)
	}
}

func Infof(format string, v ...any) {
	if defaultLogger.logLevel >= INFO {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(INFO, callerDepth, msg)
	}
}

func Warn(v ...any) {
	if defaultLogger.logLevel >= WARN {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(WARN, callerDepth, msg)
	}
}

func Warnf(format string, v ...any) {
	if defaultLogger.logLevel >= WARN {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(WARN, callerDepth, msg)
	}
}

func Error(v ...any) {
	if defaultLogger.logLevel >= ERROR {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(ERROR, callerDepth, msg)
	}
}

func Errorf(format string, v ...any) {
	if defaultLogger.logLevel >= ERROR {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(ERROR, callerDepth, msg)
	}
}

func Fatal(v ...any) {
	if defaultLogger.logLevel >= FATAL {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(FATAL, callerDepth, msg)
	}
}

func Fatalf(format string, v ...any) {
	if defaultLogger.logLevel >= FATAL {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(FATAL, callerDepth, msg)
	}
}
