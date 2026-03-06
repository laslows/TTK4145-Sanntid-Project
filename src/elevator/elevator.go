package elevator

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

//Should tidy up this file a lot. Maybe separate the get/set-functions, the driver functions and
// the smart functions

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
	//Can maybe remove IP
	m_ID        int
	m_floor     int
	m_direction Direction
	m_requests  [config.N_FLOORS][config.N_BUTTONS]bool
	m_behaviour ElevatorBehaviour
	m_isMaster  bool
	m_myBackup  *Backup

	m_worldView [config.N_ELEVATORS]*Backup

	m_Config struct {
		m_doorOpenDuration time.Duration
	}
}

// Constructor
func New(port string) *Elevator {
	e := &Elevator{
		m_floor:     -1,
		m_direction: Stop,
		m_behaviour: Idle,
		m_isMaster:  true,

		m_worldView: [config.N_ELEVATORS]*Backup{},

		m_Config: struct {
			m_doorOpenDuration time.Duration
		}{
			m_doorOpenDuration: 3 * time.Second,
		},
	}

	e.m_ID = getIDAsInt(getLocalIP(), port)

	e.m_myBackup = &Backup{
		m_ID:                 e.m_ID,
		m_floor:              e.m_floor,
		m_direction:          e.m_direction,
		m_isMaster:           e.m_isMaster,
		m_version:            0,
		m_behaviour:          Idle,
		m_connectedToNetwork: true,
	}

	e.UpdateWorldView(e.m_myBackup)

	return e
}

func (e *Elevator) GetGlobalLights() [config.N_FLOORS][config.N_BUTTONS]bool {
	lights := e.m_requests

	for _, b := range e.m_worldView {
		if b != nil {
			for f := 0; f < config.N_FLOORS; f++ {
				for btn := 0; btn < 2; btn++ {
					// Local elevator should not turn on global cab lights
					lights[f][btn] = lights[f][btn] || b.m_requests[f][btn]
				}
			}
		}
	}

	return lights
}

// Maybe this is all we need, and we dont need a function that cheks if new backup == old backup
// Should maybe use a message id instead, to check if we have already received the message
func (e *Elevator) UpdateWorldView(backup *Backup) {
	for i, b := range e.m_worldView {
		if b == nil || (b.m_ID == backup.m_ID) {
			e.m_worldView[i] = backup
			return
		}
	}
}

func (e *Elevator) TryUpdateWorldView(backup *Backup) bool {
	// Update if new elevator, or if the incoming backup is newer, or if backup has reconnected.

	for _, b := range e.m_worldView {
		if b != nil && b.m_ID == backup.m_ID {
			return backup.m_version > b.m_version || !b.m_connectedToNetwork
		}
	}
	return true
}

func getIDAsInt(ip, osID string) int {
	ipString := strings.ReplaceAll(ip, ".", "")
	iDString := ipString + osID
	idInt, err := strconv.Atoi(iDString)

	if err != nil {
		fmt.Println(iDString)
		panic(fmt.Sprintf("Failed to convert IP to int: %v", err))
	}

	return idInt
}

func (e *Elevator) TryUpdateIsMaster() bool {
	//If we are master and should be slave, or if we are slave and should be master,
	// update isMaster and return true. Else return false
	if (e.m_isMaster && !checkIsMaster(*e)) || (!e.m_isMaster && checkIsMaster(*e)) {
		e.m_isMaster = checkIsMaster(*e)
		return true
	}

	return false

}

func checkIsMaster(e Elevator) bool {
	master := true

	for _, b := range e.m_worldView {
		if b != nil && b.m_connectedToNetwork {
			master = master && (e.GetID() >= b.GetID())
		}
	}

	//Remove this lol
	if master {
		fmt.Println("I am master!")
	}

	return master
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

func (e *Elevator) GetMyBackup() *Backup {
	//Would maybe be easier to store a pointer to own backup in elevator struct, and update it every time we update the worldview

	for _, b := range e.m_worldView {
		if b != nil && b.m_ID == e.m_ID {
			return b
		}
	}
	return nil
}

func (e *Elevator) GetMasterID() int {
	for _, b := range e.m_worldView {
		if b != nil && b.m_isMaster {
			return b.GetID()
		}
	}

	fmt.Println("No master found in worldview")
	return -1
}

func (e *Elevator) LooseConnectionToPeer(peer int) {
	for i, b := range e.m_worldView {
		if b != nil && b.m_ID == peer {
			e.m_worldView[i].m_connectedToNetwork = false
			return
		}
	}
}

func (e *Elevator) GetWorldView() [config.N_ELEVATORS]*Backup {
	return e.m_worldView
}

func (e *Elevator) GetFloor() int {
	return e.m_floor
}

func (e *Elevator) SetFloor(f int) {
	e.m_floor = f
}

func (e *Elevator) GetRequestAtFloor(floor int, btn driver.ButtonType) bool {
	return e.m_requests[floor][btn]
}

func (e *Elevator) GetRequests() [config.N_FLOORS][config.N_BUTTONS]bool {
	return e.m_requests
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

func (e *Elevator) GetID() int {
	return e.m_ID
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

func DirectionToString(dir Direction) string {
	switch dir {
	case Down:
		return "down"
	case Up:
		return "up"
	case Stop:
		return "stop"
	default:
		return ""
	}
}

func BehaviourToString(behaviour ElevatorBehaviour) string {
	switch behaviour {
	case Idle:
		return "idle"
	case DoorOpen:
		return "doorOpen"
	case Moving:
		return "moving"
	case MotorStop:
		return "motorStop"
	default:
		return ""
	}
}
