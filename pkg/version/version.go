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
// Create: 2021-04-25
// Description: version releated

// Package version is for version check
package version

import (
	"fmt"
	"os"
	"runtime"
)

var (
	// Version represents rubik Version
	Version string
	// Release represents rubik Release number
	Release string
	// GitCommit represents git commit number
	GitCommit string
	// BuildTime represents build time
	BuildTime string
)

func init() {
	const maxArgLen = 2
	if len(os.Args) == maxArgLen && os.Args[1] == "-v" {
		fmt.Println("Version:      ", Version)
		fmt.Println("Release:      ", Release)
		fmt.Println("Go Version:   ", runtime.Version())
		fmt.Println("Git Commit:   ", GitCommit)
		fmt.Println("Built:        ", BuildTime)
		fmt.Println("OS/Arch:      ", runtime.GOOS+"/"+runtime.GOARCH)
		os.Exit(0)
	}
}
