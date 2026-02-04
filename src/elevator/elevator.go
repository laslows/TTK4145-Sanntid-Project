package elevator

import "./driver"

//TODO: Fix naming conventions

const N_FLOORS int = 4
const N_BUTTONS int = 3

type Dirn int

const (
	DIRN_DOWN Dirn = -1
	DIRN_STOP      = 0
	DIRN_UP        = 1
)

type Button int

const (
	BUTTON_HALL_UP Button = iota
	BUTTON_HALL_DOWN
	BUTTON_CAB
)

type ElevatorBehaviour int

const (
	EB_IDLE ElevatorBehaviour = iota
	EB_DOOR_OPEN
	EB_MOVING
)

type Elevator struct {
	Floor     int
	Dirn      Dirn
	Requests  [N_FLOORS][N_BUTTONS]int
	Behaviour ElevatorBehaviour

	Config struct {
		doorOpenDuration_s float64
	}
}

func Elevator_uninitialized() Elevator {
	elevio.Init("localhost:15657", N_FLOORS)
	return Elevator{
		Floor:     -1,
		Dirn:      DIRN_STOP,
		Behaviour: EB_IDLE,
		Config: struct {
			doorOpenDuration_s float64
		}{
			doorOpenDuration_s: 3.0,
		},
	}
}

func Test() int {
	return elevio.GetFloor()
}

func Elevator_floorSensor() int {
	return elevio.GetFloor()
}

func Elevator_requestButton(f int, b Button) bool {
	return elevio.GetButton(elevio.ButtonType(b), f)
}

func Elevator_stopButton() bool {
	return elevio.GetStop()
}

func Elevator_obstruction() bool {
	return elevio.GetObstruction()
}

func Elevator_floorIndicator(f int) {
	elevio.SetFloorIndicator(f)
}

func Elevator_requestButtonLight(f int, b Button, on bool) {
	elevio.SetButtonLamp(elevio.ButtonType(b), f, on)
}

func Elevator_doorOpenLight(on bool) {
	elevio.SetDoorOpenLamp(on)
}

func Elevator_stopButtonLight(on bool) {
	elevio.SetStopLamp(on)
}

func Elevator_motorDirection(dirn Dirn) {
	elevio.SetMotorDirection(elevio.MotorDirection(dirn))
}
