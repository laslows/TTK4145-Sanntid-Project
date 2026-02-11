package main

import (
	"./src/fsm"
	initialize "./src/init"
	"./src/network"
)

func main() {
	initialize.Initialize()

	go fsm.Fsm()
	go network.Network()

}
