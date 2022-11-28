package logger

import (
    "log"
    "os"
)

type Logger struct {
    err  *log.Logger
    warn *log.Logger
    info *log.Logger
}

func New() Logger {
    return Logger{
        err:  log.New(os.Stderr, "[ERROR]  ", log.LstdFlags),
        warn: log.New(os.Stdout, "[WARNING]", log.LstdFlags),
        info: log.New(os.Stdout, "[INFO]   ", log.LstdFlags),
    }
}

func (logger *Logger) Warning(err error) {
    logger.warn.Println(err.Error())
}

func (logger *Logger) Error(err error) {
    logger.err.Println(err.Error())
}

func (logger *Logger) InfoF(format string, args ...any) {
    logger.info.Printf(format, args...)
}

func (logger *Logger) WarningF(format string, args ...any) {
    logger.warn.Printf(format, args...)
}

func (logger *Logger) ErrorF(format string, args ...any) {
    logger.err.Printf(format, args...)
}