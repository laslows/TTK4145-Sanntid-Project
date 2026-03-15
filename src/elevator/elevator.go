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

//TODO: differ between backup and b with better namings..
//TODO: maybe mutex?

type Direction int

const (
	Down Direction = iota - 1
	Stop
	Up
)

type ElevatorBehaviour int

const (
	Idle ElevatorBehaviour = iota
	DoorOpen
	Moving
	MotorStop
)

type Elevator struct {
	m_ID        int
	m_floor     int
	m_direction Direction
	m_requests  [config.N_FLOORS][config.N_BUTTONS]bool
	m_behaviour ElevatorBehaviour
	m_isMaster  bool
	m_isObstructed bool
	m_myBackup  *Backup

	m_worldView [config.N_ELEVATORS]*Backup

	m_Config struct {
		m_doorOpenDuration time.Duration
	}
}

func New(port string) *Elevator {
	e := &Elevator{
		m_floor:     -1,
		m_direction: Down,
		m_behaviour: Idle,
		m_isMaster:  true,
		m_isObstructed: false,

		m_worldView: [config.N_ELEVATORS]*Backup{},

		m_Config: struct {
			m_doorOpenDuration time.Duration
		}{
			m_doorOpenDuration: time.Duration(config.DOOR_OPEN_DURATION) * time.Second,
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
		m_isObstructed:       false,
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

//TODO: move to fsm?
func (e *Elevator) ShouldRedistributeOrders(backup *Backup) bool {
	//SHould redistribute if new backup changes obstruction status, or if we lose connection or if we gain connection, or if we change motorstopstatus
    for _, b := range e.m_worldView {
		if b != nil && b.m_ID == backup.m_ID {
			return (b.m_isObstructed != backup.m_isObstructed || b.GetHasMotorstop() != backup.GetHasMotorstop())
		}
	}
	return false
}

func (e *Elevator) TryUpdateIsMaster() bool {
	shouldBeMaster := checkIsMaster(*e)
	if e.m_isMaster != shouldBeMaster {
		e.m_isMaster = shouldBeMaster
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

//TODO: maybe not return pointer.. Whuuups
//TODO: fix whole weird backup/worldview thing. Mybackup-pointer should be
//same as pointer in worldview
func (e *Elevator) GetMyBackup() *Backup {

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

func (e *Elevator) LoseConnectionToPeer(peerID int) {
	for i, b := range e.m_worldView {
		if b != nil && b.m_ID == peerID && e.m_ID != peerID {
			e.m_worldView[i].m_connectedToNetwork = false
			return
		}
	}
}

func (e *Elevator) RestoreElevatorState(b *Backup) {

	e.m_requests = b.m_requests
	e.m_floor = b.m_floor
	e.m_direction = b.m_direction

	e.restoreMyBackup(b)

}

func (e *Elevator) ClearDisconnectedNodeQueue(){
	for _, b := range e.m_worldView {
		if b != nil && !b.m_connectedToNetwork {
			for f := 0; f < config.N_FLOORS; f++ {
				for btn := 0; btn < config.N_BUTTONS-1; btn++ {
					b.m_requests[f][btn] = false
				}
			}
		}
	}
}

func (e *Elevator) SetIsObstructed(isObstructed bool) {
	e.m_isObstructed = isObstructed
}

//TODO: Maybe not return pointers
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
