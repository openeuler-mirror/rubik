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

type level uint32

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
	logDebug level = iota
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
func InitConfig(driver, logdir, lvl string, size int64) error {
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

	if lvl == "" {
		lvl = constant.LogLevelInfo
	}
	levelstr, err := levelFromString(lvl)
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

func levelToString(lvl level) string {
	switch lvl {
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

func levelFromString(lvl string) (level, error) {
	switch lvl {
	case constant.LogLevelDebug:
		return logDebug, nil
	case constant.LogLevelInfo, "":
		return logInfo, nil
	case constant.LogLevelWarn:
		return logWarn, nil
	case constant.LogLevelError:
		return logError, nil
	default:
		return logInfo, fmt.Errorf("invalid log level: %s", lvl)
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

func output(lvl string, format string, args ...interface{}) {
	tag := fmt.Sprintf("%s [rubik] level=%s ", time.Now().Format("2006-01-02 15:04:05.000"), lvl)
	raw := fmt.Sprintf(format, args...) + "\n"

	depth := 1
	if lvl == constant.LogLevelStack {
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
		} else if lvl == constant.LogLevelStack {
			break
		}
		writeLine(line)
	}
}

func logln(lvl level, format string, args ...interface{}) {
	if lvl >= logLevel {
		output(levelToString(lvl), format, args...)
	}
}

// Debugf output debug level logs when then log level of the logger is less than or equal to debug level
func Debugf(format string, args ...interface{}) {
	logln(logDebug, format, args...)
}

// Infof output info level logs when then log level of the logger is less than or equal to info level
func Infof(format string, args ...interface{}) {
	logln(logInfo, format, args...)
}

// Warnf output warn level logs when then log level of the logger is less than or equal to warn level
func Warnf(format string, args ...interface{}) {
	logln(logWarn, format, args...)
}

// Errorf output warn level logs when then log level of the logger is less than or equal to error level
func Errorf(format string, args ...interface{}) {
	logln(logError, format, args...)
}
