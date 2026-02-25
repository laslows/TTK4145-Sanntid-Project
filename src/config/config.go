package config

import "time"

const DOOR_OPEN_DURATION float64 = 3.0
const INPUT_POLL_RATE int = 25
const MOTOR_STOP_TIMEOUT float64 = 5 //idk
const INCLUDE_CAB bool = false
const TRAVEL_DURATION time.Duration = 0 //fix

const N_FLOORS int = 4
const N_BUTTONS int = 3

const N_ELEVATORS int = 3

type ClearRequestType int

const (
	All ClearRequestType = iota
	InDirn
)

var clearRequestType ClearRequestType = InDirn
