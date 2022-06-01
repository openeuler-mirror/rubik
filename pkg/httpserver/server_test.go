// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2021-05-12
// Description: server test case

package httpserver

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
)

type rootHandlerTC struct {
	name, method, url string
	data              []byte
	wantErr           bool
}

func createLogTC() []rootHandlerTC {
	return []rootHandlerTC{
		{
			name:    "TC1-invalid url",
			method:  "POST",
			url:     "/a",
			data:    []byte("{}"),
			wantErr: true,
		},
		{
			name:    "TC2-invalid method",
			method:  "GET",
			url:     "/",
			data:    []byte("{}"),
			wantErr: true,
		},
		{
			name:   "TC3-set qos failed",
			method: "POST",
			url:    "/",
			data: []byte("{\"Pods\": {\"pod70f2828b-3f9c-42e2-97da-01c6072af4a6\": {\"CgroupPath\": " +
				"\"kubepods/besteffort/pod70f2828b-3f9c-42e2-97da-01c6072af4a6\", \"QoSLevel\": 0 }}}"),
			wantErr: false,
		},
		{
			name:    "TC4-invalid data",
			method:  "POST",
			url:     "/",
			data:    []byte(":"),
			wantErr: false,
		},
		{
			name:    "TC5-null data",
			method:  "POST",
			url:     "/",
			data:    []byte("{}"),
			wantErr: false,
		},
		{
			name:    "TC6-ShutdownFlag true",
			method:  "POST",
			url:     "/",
			data:    []byte("{}"),
			wantErr: false,
		},
	}
}

// TestRootHandler is RootHandler function test
func TestRootHandler(t *testing.T) {
	handler := setupHandler()
	tests := createLogTC()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "TC6-ShutdownFlag true" {
				atomic.AddInt32(&config.ShutdownFlag, 1)
			}
			r, err := http.NewRequest(tt.method, tt.url, bufio.NewReader(bytes.NewReader(tt.data)))
			if err != nil {
				t.Fatal(err)
			}
			r.ContentLength = int64(len(tt.data))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			assert.Equal(t, tt.wantErr, w.Code != http.StatusOK)
		})
	}
}

// TestRootHandlerInvalidLength test invalid length
func TestRootHandlerInvalidRequest(t *testing.T) {
	handler := setupHandler()
	data := []byte("abc")
	r, err := http.NewRequest("POST", "/", bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		t.Fatal(err)
	}
	r.ContentLength = int64(len(data) + 1)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	r.ContentLength = int64(len(data) - 1)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRootHandler is PingHandler function test
func TestPingHandler(t *testing.T) {
	r, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler := setupHandler()
	handler.ServeHTTP(w, r)
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	t.Logf("http response: %s\n", w.Body.String())

	r, err = http.NewRequest("POST", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, true, w.Code != http.StatusOK)
}

// // TestRootHandler is VersionHandler function test
func TestVersionHandler(t *testing.T) {
	r, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler := setupHandler()
	handler.ServeHTTP(w, r)
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	t.Logf("http response: %s\n", w.Body.String())

	r, err = http.NewRequest("POST", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, true, w.Code != http.StatusOK)
}

// TestNewSock is NewSock function test
func TestNewSock(t *testing.T) {
	unix.Umask(constant.DefaultUmask)
	_, err := NewSock()
	assert.NoError(t, err)
	sock, err := os.Stat(constant.RubikSock)
	assert.NoError(t, err)
	sockFolder, err := os.Stat(filepath.Dir(constant.RubikSock))
	assert.NoError(t, err)
	assert.Equal(t, constant.DefaultFileMode.Perm().String(), sock.Mode().Perm().String())
	assert.Equal(t, constant.DefaultDirMode.Perm().String(), sockFolder.Mode().Perm().String())
	err = os.RemoveAll(constant.RubikSock)
	assert.NoError(t, err)
}

// TestNewServer is NewServer function test
func TestNewServer(t *testing.T) {
	server, pool := NewServer()
	assert.Equal(t, constant.ReadTimeout, server.ReadTimeout)
	assert.Equal(t, constant.WriteTimeout, server.WriteTimeout)
	assert.Equal(t, constant.WorkerNum, pool.WorkerNum)
}
