// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: rubik team
// Create: 2021-04-17
// Description: This file is used for unix socket server

package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/go-errors/errors"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/version"
	"isula.org/rubik/pkg/workerpool"
)

const (
	handleTimeout = 2 * time.Minute
)

var pool *workerpool.WorkerPool

// NewSock creates a new rubik sock
func NewSock() (*net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(constant.RubikSock), constant.DefaultDirMode); err != nil {
		return nil, err
	}
	if err := os.Remove(constant.RubikSock); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	sock, err := net.Listen("unix", constant.RubikSock)
	if err != nil {
		return nil, err
	}

	if err = os.Chmod(constant.RubikSock, constant.DefaultFileMode); err != nil {
		return nil, err
	}

	return &sock, nil
}

// NewServer creates a new http server
func NewServer() (*http.Server, *workerpool.WorkerPool) {
	server := &http.Server{
		ReadTimeout:  constant.ReadTimeout,
		WriteTimeout: constant.WriteTimeout,
		Handler:      setupHandler(),
	}

	return server, pool
}

func newContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	uuid := time.Now().Format("150405.999999")
	ctx = context.WithValue(ctx, log.CtxKey(log.UUID), uuid)
	return ctx, cancel
}

// RootHandler is used for processing POST request to setting qos level for pods
func RootHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := newContext(handleTimeout)
	defer cancel()

	b := readBody(ctx, w, r)
	if b == nil {
		return
	}

	if atomic.LoadInt32(&config.ShutdownFlag) != 0 {
		writeRootResponse(ctx, w, constant.ErrCodeFailed, "Server is shutdown, stop handle request")
		return
	}

	log.WithCtx(ctx).Logf("Handle HTTP root request start")
	if err := pool.PushTask(workerpool.NewQosTask(ctx, b)); err != nil {
		writeRootResponse(ctx, w, constant.ErrCodeFailed, "set qos failed")
		log.WithCtx(ctx).Errorf("Handle HTTP root request failed: %v", err)
		return
	}

	writeRootResponse(ctx, w, constant.DefaultSucceedCode, "")
	log.WithCtx(ctx).Logf("Handle HTTP root request OK")
}

func readBody(ctx context.Context, w http.ResponseWriter, r *http.Request) []byte {
	if r.URL.RequestURI() != "/" {
		http.NotFound(w, r)
		return nil
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return nil
	}

	maxLength := int64(1024 * 1024) // 1M
	if r.ContentLength > maxLength {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return nil
	}

	if r.ContentLength <= 0 {
		http.Error(w, http.StatusText(http.StatusLengthRequired), http.StatusLengthRequired)
		return nil
	}

	resb, err := safeRead(r.Body, int(r.ContentLength))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		log.WithCtx(ctx).Errorf("Read request body error: %v", err)
		return nil
	}

	return resb
}

func safeRead(body io.ReadCloser, size int) ([]byte, error) {
	var (
		resb []byte
		m, n int
		err  error
	)

	for {
		const minRead = 512
		b := make([]byte, minRead)
		m, err = body.Read(b)
		n += m
		if n > size {
			break
		}
		resb = append(resb, b[:m]...)
		if err != nil || err == io.EOF {
			break
		}
	}
	defer body.Close()

	if err != nil && err != io.EOF {
		return nil, err
	}
	if n > size {
		return nil, errors.Errorf("request content actual length %v larger than defined length %v", n, size)
	}
	if n < size {
		return nil, errors.Errorf("request content actual length %v shorter than defined length %v", n, size)
	}

	return resb, nil
}

func ping(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	writeResponse(ctx, w, []byte("ok"))
}

// PingHandler is used for check if rubik is still alive or not
func PingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := newContext(handleTimeout)
	ping(ctx, w, r)
	cancel()
}

func versionHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	msg, err := json.Marshal(api.VersionResponse{Version: version.Version, Release: version.Release,
		GitCommit: version.GitCommit, BuildTime: version.BuildTime, Usage: version.Usage})
	log.DropError(err)
	writeResponse(ctx, w, msg)
}

// VersionHandler is used for check if rubik version
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := newContext(handleTimeout)
	versionHandler(ctx, w, r)
	cancel()
}

func writeResponse(ctx context.Context, w http.ResponseWriter, data []byte) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.WithCtx(ctx).Errorf("Write http response failed: %v", err.Error())
	}
}

func writeRootResponse(ctx context.Context, w http.ResponseWriter, errCode int, msg string) {
	data, err := json.Marshal(api.SetQosResponse{ErrCode: errCode, Message: msg})
	if err != nil {
		log.WithCtx(ctx).Errorf("encode HTTP response failed")
	}
	writeResponse(ctx, w, data)
}

// setupHandler defines handlers and do some pre-setting
func setupHandler() *http.ServeMux {
	pool = workerpool.NewWorkerPool(constant.WorkerNum, constant.TaskChanCapacity)
	pool.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/", RootHandler)
	mux.HandleFunc("/ping", PingHandler)
	mux.HandleFunc("/version", VersionHandler)

	return mux
}
