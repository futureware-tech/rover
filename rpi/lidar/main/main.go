package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dasfoo/rover/rpi/lidar"
)

func continuous(lidar *lidar.Lidar, maxNumberOfResults int) {
	for i := 0; i < maxNumberOfResults; i++ {
		time.Sleep(1 * time.Second)
		val, e := lidar.Distance(false)
		if e == nil {
			fmt.Println(val)
		} else {
			log.Println(e)
		}
	}

}

func main() {
	log.SetFlags(log.Lshortfile)
	lidar := lidar.NewLidar(1, 0x62) // 0x62 the default LidarLite address
	defer lidar.Close()
	command := os.Args[1]                               //TODO Error check
	maxNumberOfResults, err := strconv.Atoi(os.Args[2]) //TODO Error check
	if err != nil {
		log.Println(err)
		return
	}
	switch {
	case command == "distance":
		for i := 0; i < maxNumberOfResults; i++ {
			if val, err := lidar.Distance(false); err == nil {
				fmt.Println(val)
				time.Sleep(1 * time.Second)
			}
		}
	case command == "velocity":
		for i := 0; i < maxNumberOfResults; i++ {
			if val, err := lidar.Velocity(); err == nil {
				fmt.Println(val)
				time.Sleep(1 * time.Second)
			}
		}
	case command == "continuous":
		if err := lidar.BeginContinuous(true, 0xc8, 0xff); err == nil {
			continuous(lidar, maxNumberOfResults)
		} else {
			log.Println(err)
		}
		// For testing purpose
	case command == "test":
		if err := lidar.BeginContinuous(true, 0xc8, 0xff); err == nil {
			continuous(lidar, maxNumberOfResults)
			lidar.StopContinuous()
			for i := 0; i < maxNumberOfResults; i++ {
				if val, e := lidar.Distance(false); e == nil {
					fmt.Println(val)
					time.Sleep(1 * time.Second)
				}
			}
		} else {
			log.Println(err)
		}
	}
}
