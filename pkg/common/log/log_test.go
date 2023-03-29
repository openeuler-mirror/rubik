// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2021-05-24
// Description: This file is used for testing tinylog

package log

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/test/try"
)

// test_rubik_set_logdriver_0001
func TestInitConfigLogDriver(t *testing.T) {
	logDir := try.GenTestDir().String()
	logFilePath := filepath.Join(logDir, "rubik.log")

	// case: rubik.log already exist.
	try.WriteFile(logFilePath, "")
	err := InitConfig("file", logDir, "", logSize)
	assert.NoError(t, err)

	err = os.RemoveAll(logDir)
	assert.NoError(t, err)

	// logDriver is file
	err = InitConfig("file", logDir, "", logSize)
	assert.NoError(t, err)
	assert.Equal(t, file, logDriver)
	logString := "Test InitConfig with logDriver file"
	Infof(logString)
	b, err := ioutil.ReadFile(logFilePath)
	assert.NoError(t, err)
	assert.Equal(t, true, strings.Contains(string(b), logString))

	// logDriver is stdio
	os.Remove(logFilePath)
	err = InitConfig("stdio", logDir, "", logSize)
	assert.NoError(t, err)
	assert.Equal(t, stdio, logDriver)
	logString = "Test InitConfig with logDriver stdio"
	Infof(logString)
	_, err = ioutil.ReadFile(logFilePath)
	assert.Equal(t, true, err != nil)

	// logDriver invalid
	err = InitConfig("std", logDir, "", logSize)
	assert.Equal(t, true, err != nil)

	// logDriver is null
	err = InitConfig("", logDir, "", logSize)
	assert.NoError(t, err)
	assert.Equal(t, stdio, logDriver)
}

// test_rubik_set_logdir_0001
func TestInitConfigLogDir(t *testing.T) {
	logDir := try.GenTestDir().String()
	logFilePath := filepath.Join(logDir, "rubik.log")

	// LogDir valid
	err := InitConfig("file", logDir, "", logSize)
	assert.NoError(t, err)
	logString := "Test InitConfig with logDir valid"
	Infof(logString)
	b, err := ioutil.ReadFile(logFilePath)
	assert.NoError(t, err)
	assert.Equal(t, true, strings.Contains(string(b), logString))

	// logDir invalid
	err = InitConfig("file", "invalid/log", "", logSize)
	assert.Equal(t, true, err != nil)
}

type logTC struct {
	name, logLevel              string
	wantErr, debug, info, error bool
}

func createLogTC() []logTC {
	return []logTC{
		{
			name:     "TC1-logLevel debug",
			logLevel: "debug",
			wantErr:  false,
			debug:    true,
			info:     true,
			error:    true,
		},
		{
			name:     "TC2-logLevel info",
			logLevel: "info",
			wantErr:  false,
			debug:    false,
			info:     true,
			error:    true,
		},
		{
			name:     "TC3-logLevel error",
			logLevel: "error",
			wantErr:  false,
			debug:    false,
			info:     false,
			error:    true,
		},
		{
			name:     "TC4-logLevel null",
			logLevel: "",
			wantErr:  false,
			debug:    false,
			info:     true,
			error:    true,
		},
		{
			name:     "TC5-logLevel invalid",
			logLevel: "inf",
			wantErr:  true,
		},
	}
}

// test_rubik_set_loglevel_0001
func TestInitConfigLogLevel(t *testing.T) {
	logDir := try.GenTestDir().String()
	logFilePath := filepath.Join(logDir, "rubik.log")

	debugLogSting, infoLogSting, errorLogSting, logLogString := "Test InitConfig debug log",
		"Test InitConfig info log", "Test InitConfig error log", "Test InitConfig log log"
	for _, tt := range createLogTC() {
		t.Run(tt.name, func(t *testing.T) {
			err := InitConfig("file", logDir, tt.logLevel, logSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitConfig() = %v, want %v", err, tt.wantErr)
			} else if tt.wantErr == false {
				Debugf(debugLogSting)
				Infof(infoLogSting)
				Errorf(errorLogSting)
				Infof(logLogString)
				b, err := ioutil.ReadFile(logFilePath)
				assert.NoError(t, err)
				assert.Equal(t, tt.debug, strings.Contains(string(b), debugLogSting))
				assert.Equal(t, tt.info, strings.Contains(string(b), infoLogSting))
				assert.Equal(t, tt.info, strings.Contains(string(b), logLogString))
				assert.Equal(t, tt.error, strings.Contains(string(b), errorLogSting))
				os.Remove(logFilePath)

				ctx := context.WithValue(context.Background(), CtxKey(constant.LogEntryKey), "abc123")
				WithCtx(ctx).Debugf(debugLogSting)
				WithCtx(ctx).Infof(infoLogSting)
				WithCtx(ctx).Errorf(errorLogSting)
				WithCtx(ctx).Warnf(logLogString)
				b, err = ioutil.ReadFile(logFilePath)
				assert.NoError(t, err)
				assert.Equal(t, tt.debug, strings.Contains(string(b), debugLogSting))
				assert.Equal(t, tt.info, strings.Contains(string(b), infoLogSting))
				assert.Equal(t, tt.error, strings.Contains(string(b), errorLogSting))
				assert.Equal(t, tt.info, strings.Contains(string(b), logLogString))
				assert.Equal(t, true, strings.Contains(string(b), "abc123"))
				err = os.RemoveAll(logDir)
				assert.NoError(t, err)
			}
		})
	}
}

// test_rubik_set_logsize_0001
func TestInitConfigLogSize(t *testing.T) {
	logDir := try.GenTestDir().String()
	// LogSize invalid
	err := InitConfig("file", logDir, "", logSizeMin-1)
	assert.Equal(t, true, err != nil)
	err = InitConfig("file", logDir, "", logSizeMax+1)
	assert.Equal(t, true, err != nil)

	// logSize valid
	testSize, printLine, repeat := 100, 50000, 100
	err = InitConfig("file", logDir, "", logSize)
	assert.NoError(t, err)
	for i := 0; i < printLine; i++ {
		Infof(strings.Repeat("TestInitConfigLogSize log", repeat))
	}
	err = InitConfig("file", logDir, "", int64(testSize))
	assert.NoError(t, err)
	for i := 0; i < printLine; i++ {
		Infof(strings.Repeat("TestInitConfigLogSize log", repeat))
	}
	var size int64
	err = filepath.Walk(logDir, func(_ string, f os.FileInfo, _ error) error {
		size += f.Size()
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, true, size < int64(testSize)*unitMB)
	err = os.RemoveAll(constant.TmpTestDir)
	assert.NoError(t, err)
}

// TestLogStack is Stackf function test
func TestLogStack(t *testing.T) {
	logDir := try.GenTestDir().String()
	logFilePath := filepath.Join(logDir, "rubik.log")

	err := InitConfig("file", logDir, "", logSize)
	assert.NoError(t, err)
	Stackf("test stack log")
	b, err := ioutil.ReadFile(logFilePath)
	assert.NoError(t, err)
	fmt.Println(string(b))
	assert.Equal(t, true, strings.Contains(string(b), t.Name()))
	line := strings.Split(string(b), "\n")
	maxLineNum := 5
	assert.Equal(t, true, len(line) < maxLineNum)
}

// TestDropError is DropError function test
func TestDropError(t *testing.T) {
	logDir := try.GenTestDir().String()
	logFilePath := filepath.Join(logDir, "rubik.log")

	err := InitConfig("file", logDir, "", logSize)
	assert.NoError(t, err)
	DropError()
	dropError := "test drop error"
	DropError(dropError)
	DropError(nil)
	_, err = ioutil.ReadFile(logFilePath)
	assert.Equal(t, true, err != nil)
}

// TestLogOthers is log other tests
func TestLogOthers(t *testing.T) {
	logDir := filepath.Join(try.GenTestDir().String(), "regular-file")
	try.WriteFile(logDir, "")

	err := makeLogDir(logDir)
	assert.Equal(t, true, err != nil)

	const outOfRangeLogLevel = 100
	s := levelToString(outOfRangeLogLevel)
	assert.Equal(t, "", s)
	const stackLoglevel = 20
	s = levelToString(stackLoglevel)
	assert.Equal(t, "stack", s)

	logDriver = 1
	logFname = filepath.Join(constant.TmpTestDir, "log-not-exist")
	os.MkdirAll(logFname, constant.DefaultDirMode)
	writeLine("abc")

	s = WithCtx(context.Background()).level(1)
	assert.Equal(t, "info", s)

	logLevel = logError + 1
	WithCtx(context.Background()).Errorf("abc")
}
