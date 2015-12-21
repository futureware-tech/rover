package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dasfoo/rover/rpi/lidar"
)

func main() {
	log.SetFlags(log.Lshortfile)
	lidar := lidar.NewLidar(1, 0x62) // 0x62 the default LidarLite address
	defer lidar.CloseLidar()

	for {
		//if val, err := lidar.Distance(false); err == nil {
		//	fmt.Println(val)
		//	time.Sleep(1 * time.Second)
		//}

		if val, err := lidar.Velocity(); err == nil {
			fmt.Println(val)
			time.Sleep(1 * time.Second)
		}
	}

	/*if err := lidar.BeginContinuous(true, 0x08, 0xff); err == nil {
		for {
			val, e := lidar.DistanceContinuous()

			if e == nil {
				fmt.Println(val)
			}
		}
	}*/
}
