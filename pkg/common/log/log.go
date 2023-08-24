// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Haomin Tsai
// Create: 2021-09-28
// Description: This file is used for rubik log

// Package tinylog is for rubik log
package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"isula.org/rubik/pkg/common/constant"
)

// CtxKey used for UUID
type CtxKey string

const (
	logStack           = 20
	logStackFrom       = 2
	logFileNum         = 10
	logSizeMin   int64 = 10          // 10MB
	logSizeMax   int64 = 1024 * 1024 // 1TB
	unitMB       int64 = 1024 * 1024
)

const (
	stdio int = iota
	file
)

const (
	logDebug int = iota
	logInfo
	logWarn
	logError
)

var (
	logDriver            = stdio
	logFname             = filepath.Join(constant.DefaultLogDir, "rubik.log")
	logLevel             = logInfo
	logSize        int64 = 1024
	logFileMaxSize int64
	logFileSize    int64
	lock           = sync.Mutex{}
)

func makeLogDir(logDir string) error {
	if !filepath.IsAbs(logDir) {
		return fmt.Errorf("invalid path, log directory must be an absolute path: %v", logDir)
	}

	if err := os.MkdirAll(logDir, constant.DefaultDirMode); err != nil {
		return fmt.Errorf("failed to create log directory %v : %v", logDir, err)
	}

	return nil
}

// InitConfig initializes log config
func InitConfig(driver, logdir, level string, size int64) error {
	if driver == "" {
		driver = constant.LogDriverStdio
	}
	if driver != constant.LogDriverStdio && driver != constant.LogDriverFile {
		return fmt.Errorf("invalid log driver: %s", driver)
	}
	logDriver = stdio
	if driver == constant.LogDriverFile {
		logDriver = file
	}

	if level == "" {
		level = constant.LogLevelInfo
	}
	levelstr, err := levelFromString(level)
	if err != nil {
		return err
	}
	logLevel = levelstr

	if size < logSizeMin || size > logSizeMax {
		return fmt.Errorf("invalid log size: %d (valid range is %d-%d)", size, logSizeMin, logSizeMax)
	}
	logSize = size
	logFileMaxSize = logSize / logFileNum

	if driver == constant.LogDriverFile {
		if err := makeLogDir(logdir); err != nil {
			return err
		}
		logFname = filepath.Join(logdir, "rubik.log")
		if f, err := os.Stat(logFname); err == nil {
			atomic.StoreInt64(&logFileSize, f.Size())
		}
	}

	return nil
}

// DropError drop unused error
func DropError(args ...interface{}) {
	argn := len(args)
	if argn == 0 {
		return
	}
	arg := args[argn-1]
	if arg != nil {
		fmt.Printf("drop error: %v\n", arg)
	}
}

func levelToString(level int) string {
	switch level {
	case logDebug:
		return constant.LogLevelDebug
	case logInfo:
		return constant.LogLevelInfo
	case logWarn:
		return constant.LogLevelWarn
	case logError:
		return constant.LogLevelError
	case logStack:
		return constant.LogLevelStack
	default:
		return ""
	}
}

func levelFromString(level string) (int, error) {
	switch level {
	case constant.LogLevelDebug:
		return logDebug, nil
	case constant.LogLevelInfo, "":
		return logInfo, nil
	case constant.LogLevelWarn:
		return logWarn, nil
	case constant.LogLevelError:
		return logError, nil
	default:
		return logInfo, fmt.Errorf("invalid log level: %s", level)
	}
}

func renameLogFile() {
	// rename the log that is no longer recorded, that is, the log of rubik.log.X
	for i := logFileNum - 1; i > 1; i-- {
		oldFile := logFname + fmt.Sprintf(".%d", i-1)
		newFile := logFname + fmt.Sprintf(".%d", i)
		if _, err := os.Stat(oldFile); err == nil {
			DropError(os.Rename(oldFile, newFile))
		}
	}

	// dump the current rubik log
	firstDumpLogName := logFname + ".1"
	DropError(os.Rename(logFname, firstDumpLogName))
	// change log file permissions
	DropError(os.Chmod(firstDumpLogName, constant.DefaultDumpLogFileMode))
}

func rotateLog(line int64) string {
	if atomic.AddInt64(&logFileSize, line) > logFileMaxSize*unitMB {
		renameLogFile()
		atomic.StoreInt64(&logFileSize, line)
	}

	return logFname
}

func writeLine(line string) {
	lock.Lock()
	defer lock.Unlock()

	if logDriver == stdio {
		fmt.Printf("%s", line)
		return
	}

	f, err := os.OpenFile(rotateLog(int64(len(line))), os.O_CREATE|os.O_APPEND|os.O_WRONLY, constant.DefaultFileMode)
	if err != nil {
		return
	}

	DropError(f.WriteString(line))
	DropError(f.Close())
}

func output(level string, format string, args ...interface{}) {
	tag := fmt.Sprintf("%s [rubik] level=%s ", time.Now().Format("2006-01-02 15:04:05.000"), level)
	raw := fmt.Sprintf(format, args...) + "\n"

	depth := 1
	if level == constant.LogLevelStack {
		depth = logStack
	}

	for i := logStackFrom; i < logStackFrom+depth; i++ {
		line := tag + raw
		pc, file, linum, ok := runtime.Caller(i)
		if ok {
			fs := strings.Split(runtime.FuncForPC(pc).Name(), "/")
			fs = strings.Split("."+fs[len(fs)-1], ".")
			fn := fs[len(fs)-1]
			line = tag + fmt.Sprintf("%s:%d:%s() ", file, linum, fn) + raw
		} else if level == constant.LogLevelStack {
			break
		}
		writeLine(line)
	}
}

// Warnf log warn level
func Warnf(format string, args ...interface{}) {
	if logWarn >= logLevel {
		output(levelToString(logWarn), format, args...)
	}
}

// Infof log info level
func Infof(format string, args ...interface{}) {
	if logInfo >= logLevel {
		output(levelToString(logInfo), format, args...)
	}
}

// Debugf log debug level
func Debugf(format string, args ...interface{}) {
	if logDebug >= logLevel {
		output(levelToString(logDebug), format, args...)
	}
}

// Errorf log error level
func Errorf(format string, args ...interface{}) {
	if logError >= logLevel {
		output(levelToString(logError), format, args...)
	}
}

// Stackf log stack dump
func Stackf(format string, args ...interface{}) {
	output("stack", format, args...)
}

// Entry is log entry
type Entry struct {
	Ctx context.Context
}

// WithCtx create entry with ctx
func WithCtx(ctx context.Context) *Entry {
	return &Entry{
		Ctx: ctx,
	}
}

func (e *Entry) level(l int) string {
	id, ok := e.Ctx.Value(CtxKey(constant.LogEntryKey)).(string)
	if ok {
		return levelToString(l) + " " + constant.LogEntryKey + "=" + id
	}
	return levelToString(l)
}

// Warnf write logs
func (e *Entry) Warnf(f string, args ...interface{}) {
	if logInfo < logLevel {
		return
	}
	output(e.level(logWarn), f, args...)
}

// Infof write logs
func (e *Entry) Infof(f string, args ...interface{}) {
	if logInfo < logLevel {
		return
	}
	output(e.level(logInfo), f, args...)
}

// Debugf write verbose logs
func (e *Entry) Debugf(f string, args ...interface{}) {
	if logDebug < logLevel {
		return
	}
	output(e.level(logDebug), f, args...)
}

// Errorf write error logs
func (e *Entry) Errorf(f string, args ...interface{}) {
	if logError < logLevel {
		return
	}
	output(e.level(logError), f, args...)
}
