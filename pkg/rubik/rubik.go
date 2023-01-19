package rubik

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/services"
)

func Run() int {
	fmt.Println("rubik running")
	stopChan := make(chan struct{})
	go signalHandler(stopChan)

	c := config.NewConfig(config.JSON)
	if err := c.LoadConfig(constant.ConfigFile); err != nil {
		log.Errorf("load config failed: %v\n", err)
		return -1
	}

	if err := c.PrepareServices(); err != nil {
		log.Errorf("prepare services failed: %v\n", err)
		return -1
	}

	sm := services.GetServiceManager()
	sm.Setup()
	sm.Run(stopChan)
	select {
	case <-stopChan:
		for _, s := range sm.RunningServices {
			s.TearDown()
		}
		return 0
	}
}

func signalHandler(stopChan chan struct{}) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	for sig := range signalChan {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			log.Infof("signal %v received and starting exit...", sig)
			close(stopChan)
		}
	}
}
