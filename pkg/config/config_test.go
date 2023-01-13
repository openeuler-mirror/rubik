package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/services"
	_ "isula.org/rubik/pkg/services/blkio"
	_ "isula.org/rubik/pkg/services/cachelimit"
)

var rubikConfig string = `
{
	"agent": {
	  "logDriver": "stdio",
	  "logDir": "/var/log/rubik",
	  "logSize": 2048,
	  "logLevel": "info"
	},
	"blkio":{},
	"cacheLimit": {
	  "defaultLimitMode": "static",
	  "adjustInterval": 1000,
	  "perfDuration": 1000,
	  "l3Percent": {
		"low": 20,
		"mid": 30,
		"high": 50
	  },
	  "memBandPercent": {
		"low": 10,
		"mid": 30,
		"high": 50
	  }
	}
}
`

func TestPrepareServices(t *testing.T) {

	if !util.PathExist(constant.TmpTestDir) {
		if err := os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode); err != nil {
			assert.NoError(t, err)
		}
	}

	defer os.RemoveAll(constant.TmpTestDir)

	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	defer os.Remove(tmpConfigFile)
	if err := ioutil.WriteFile(tmpConfigFile, []byte(rubikConfig), constant.DefaultFileMode); err != nil {
		assert.NoError(t, err)
		return
	}

	c := NewConfig(JSON)
	if err := c.LoadConfig(tmpConfigFile); err != nil {
		assert.NoError(t, err)
		return
	}
	if err := c.PrepareServices(); err != nil {
		assert.NoError(t, err)
		return
	}
	fmt.Printf("agent: %v\n", c.Agent)
	for name, service := range services.GetServiceManager().RunningServices {
		fmt.Printf("name: %s, service: %v\n", name, service)
	}
	for name, service := range services.GetServiceManager().RunningPersistentServices {
		fmt.Printf("name: %s, persistent service: %v\n", name, service)
	}
}
