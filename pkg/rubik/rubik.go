package rubik

import (
	"fmt"

	// import packages to auto register service
	_ "isula.org/rubik/pkg/modules/blkio"
	_ "isula.org/rubik/pkg/modules/cachelimit"
	_ "isula.org/rubik/pkg/modules/cpu"
	_ "isula.org/rubik/pkg/modules/iocost"
	_ "isula.org/rubik/pkg/modules/memory"
	_ "isula.org/rubik/pkg/version"
)

func Run() int {
	fmt.Println("rubik running")
	return 0
}
