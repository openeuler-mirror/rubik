package util

import (
	"path/filepath"

	"isula.org/rubik/pkg/common/constant"
)

var (
	// CgroupRoot is the unique cgroup mount point globally
	CgroupRoot = constant.DefaultCgroupRoot
)

// AbsoluteCgroupPath returns absolute cgroup path of specified subsystem of a relative path
func AbsoluteCgroupPath(subsys string, relativePath string) string {
	if subsys == "" || relativePath == "" {
		return ""
	}
	return filepath.Join(CgroupRoot, subsys, relativePath)
}
