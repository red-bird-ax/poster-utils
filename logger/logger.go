package logger

import (
    "log"
    "os"
)

type Logger struct {
    err  *log.Logger
    warn *log.Logger
    name string
}

func New(name string) Logger {
    return Logger{
        err:  log.New(os.Stderr, "ERR:  ", log.LstdFlags),
        warn: log.New(os.Stdout, "WARN: ", log.LstdFlags),
        name: name,
    }
}

func (logger *Logger) Warning(err error) {
    logger.warn.Print(logger.name, " ")
    logger.warn.Println(err.Error())
}

func (logger *Logger) Error(err error) {
    logger.warn.Print(logger.name, " ")
    logger.err.Println(err.Error())
}