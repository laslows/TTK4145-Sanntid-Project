package elevator

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"fmt"
	"net"
	"time"
)

type Direction int

const (
	Down Direction = iota - 1
	Stop
	Up
)

type Button int

const (
	HallUp Button = iota
	HallDown
	Cab
)

type ElevatorBehaviour int

const (
	Idle ElevatorBehaviour = iota
	DoorOpen
	Moving
	MotorStop
)

type Elevator struct {
	m_IP        string
	m_floor     int
	m_direction Direction
	m_requests  [config.N_FLOORS][config.N_BUTTONS]bool
	m_behaviour ElevatorBehaviour
	m_isMaster  bool

	m_Config struct {
		m_doorOpenDuration time.Duration
	}
}

// Constructor
func New(port string) *Elevator {
	return &Elevator{
		m_IP:        getLocalIP() + port,
		m_floor:     -1,
		m_direction: Stop,
		m_behaviour: Idle,
		m_isMaster:  false,
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

func (e *Elevator) GetRequest(floor int, btn driver.ButtonType) bool {
	return e.m_requests[floor][btn]
}

func (e *Elevator) SetRequest(floor int, btn driver.ButtonType, active bool) {
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

func (e *Elevator) GetIsMaster() bool {
	return e.m_isMaster
}

func (e *Elevator) SetIsMaster(isMaster bool) {
	e.m_isMaster = isMaster
}

func FloorSensor() int {
	return driver.GetFloor()
}

func RequestButton(floor int, btn driver.ButtonType) bool {
	return driver.GetButton(btn, floor)
}

func StopButton() bool {
	return driver.GetStop()
}

func ObstructionSwitch() bool {
	return driver.GetObstruction()
}

func FloorIndicator(floor int) {
	driver.SetFloorIndicator(floor)
}

func RequestButtonLight(floor int, btn driver.ButtonType, on bool) {
	driver.SetButtonLamp(btn, floor, on)
}

func DoorOpenLight(on bool) {
	driver.SetDoorOpenLamp(on)
}

func StopLight(on bool) {
	driver.SetStopLamp(on)
}

func MotorDirection(dir Direction) {
	driver.SetMotorDirection(driver.MotorDirection(dir))
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic("Failed to get local IP address")
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)
	fmt.Printf("Found IP address: %s\n", localAddress.IP.String())
	return localAddress.IP.String()
}
