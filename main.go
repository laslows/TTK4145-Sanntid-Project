package main

import (
	"./src/init"
	"./src/network"
	"./src/fsm"
)

func main() {
	initialize.Initialize()

	go fsm.Fsm()
	go network.Network()

	
}


