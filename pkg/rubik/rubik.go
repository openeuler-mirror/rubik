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
// Create: 2021-05-20
// Description: This file is used for rubik struct

// Package rubik is for rubik struct
package rubik

import (
	"fmt"
	"net"
	"net/http"

	"github.com/pkg/errors"

	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/httpserver"
	"isula.org/rubik/pkg/workerpool"
)

// Rubik defines Rubik struct
type Rubik struct {
	server *http.Server
	pool   *workerpool.WorkerPool
	sock   *net.Listener
}

// NewRubik creates a new rubik object
func NewRubik() (*Rubik, error) {
	sock, err := httpserver.NewSock()
	if err != nil {
		return nil, errors.Errorf("new sock failed: %v", err)
	}
	server, pool := httpserver.NewServer()

	return &Rubik{
		server: server,
		pool:   pool,
		sock:   sock,
	}, nil
}

// Serve starts http server
func (r *Rubik) Serve() error {
	return r.server.Serve(*r.sock)
}

func run() int {
	rubik, err := NewRubik()
	if err != nil {
		fmt.Printf("new rubik failed: %v\n", err)
		return constant.ErrCodeFailed
	}

	if err = rubik.Serve(); err != nil {
		return constant.ErrCodeFailed
	}
	return 0
}

// Run start rubik server
func Run() int {
	ret := run()

	return ret
}
