// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2023-01-28
// Description: This file defines DisposableChannel and relative methods

// Package util implementes public methods
package util

import (
	"sync/atomic"
)

// DisposableChannel only accepts data once and actively closes the channel, only allowed to close once
type DisposableChannel struct {
	closeCount uint32
	ch         chan struct{}
}

// NewDisposableChannel creates and returns a DisposableChannel pointer variable
func NewDisposableChannel() *DisposableChannel {
	return &DisposableChannel{
		closeCount: 0,
		ch:         make(chan struct{}, 1),
	}
}

// Close only allows closing the channel once if there are multiple calls
func (ch *DisposableChannel) Close() {
	if atomic.AddUint32(&ch.closeCount, 1) == 1 {
		close(ch.ch)
	}
}

// Wait waits for messages until the channel is closed or a message is received
func (ch *DisposableChannel) Wait() {
	if atomic.LoadUint32(&ch.closeCount) > 0 {
		return
	}
	<-ch.ch
	ch.Close()
}

// Channel returns the channel entity if the channel is not closed
func (ch *DisposableChannel) Channel() chan struct{} {
	if atomic.LoadUint32(&ch.closeCount) == 0 {
		return ch.ch
	}
	return nil
}
