package elevator

import (
	"../Driver-go/elevio"
	"time"
)

const N_FLOORS int = 4
const N_BUTTONS int = 3

type Direction int

const (
	Dwon Direction = -1
	Stop           = 0
	Up             = 1
)

type ButtonType int

const (
	HallUp ButtonType = iota
	HallDown
	Cab
)

type ElevatorBehaviour int

const (
	Idle ElevatorBehaviour = iota
	DoorOpen
	Moving
)

type Elevator struct {
	m_floor     int
	m_direction Direction
	m_requests  [N_FLOORS][N_BUTTONS]bool
	m_behaviour ElevatorBehaviour

	m_Config struct {
		m_doorOpenDuration time.Duration
	}
}

// Constructor
func New() *Elevator {
	return &Elevator{
		m_floor:     -1,
		m_direction: Stop,
		m_behaviour: Idle,
		m_Config: struct {
			m_doorOpenDuration time.Duration
		}{
			m_doorOpenDuration: 3 * time.Second,
		},
	}
}

func (e *Elevator) GetFloor() int {
	return e.m_floor
}

func (e *Elevator) SetFloor(f int) {
	e.m_floor = f
}

func (e *Elevator) GetRequest(floor int, btn ButtonType) bool {
	return e.m_requests[floor][btn]
}

func (e *Elevator) SetRequest(floor int, btn ButtonType, active bool) {
	e.m_requests[floor][btn] = active
}

func (e *Elevator) GetDirection() Direction {
	return e.m_direction
}

func (e *Elevator) SetDirection(d Direction) {
	e.m_direction = d
}

func (e *Elevator) GetBehaviour() ElevatorBehaviour {
	return e.m_behaviour
}

func (e *Elevator) SetBehaviour(b ElevatorBehaviour) {
	e.m_behaviour = b
}

func (e *Elevator) GetDoorOpenDuration() time.Duration {
	return e.m_Config.m_doorOpenDuration
}

func FloorSensor() int {
	return elevio.GetFloor()
}

func RequestButton(floor int, btn ButtonType) bool {
	return elevio.GetButton(btn, floor)
}

func StopButton() bool {
	return elevio.GetStop()
}

func ObstructionSwitch() bool {
	return elevio.GetObstruction()
}

func FloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}

func RequestButtonLight(floor int, btn ButtonType, on bool) {
	elevio.SetButtonLamp(btn, floor, on)
}

func DoorOpenLight(on bool) {
	elevio.SetDoorOpenLamp(on)
}

func StopLight(on bool) {
	elevio.SetStopLamp(on)
}

func MotorDirection(dir Direction) {
	elevio.SetMotorDirection(MotorDirection(dir))
}
