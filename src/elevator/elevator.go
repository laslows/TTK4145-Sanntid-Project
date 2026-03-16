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
	m_myRequests  [config.N_FLOORS][config.N_BUTTONS]bool
	m_globalRequests [config.N_FLOORS][config.N_BUTTONS-1]bool
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

func(e *Elevator) SetGlobalRequest(floor int, btn driver.ButtonType, active bool){
	e.m_globalRequests[floor][btn] = active
}

// Hall lights are shared from the global hall request list, while cab lights are local.
func (e *Elevator) GetAllLights() [config.N_FLOORS][config.N_BUTTONS]bool {
	var lights [config.N_FLOORS][config.N_BUTTONS]bool

	for floor := 0; floor < config.N_FLOORS; floor++ {
		lights[floor][driver.BT_HallUp] = e.m_globalRequests[floor][driver.BT_HallUp]
		lights[floor][driver.BT_HallDown] = e.m_globalRequests[floor][driver.BT_HallDown]
		lights[floor][driver.BT_Cab] = e.m_myRequests[floor][driver.BT_Cab]
	}

	return lights
}


func (e *Elevator) UpdateWorldView(incomingBackup *Backup) {
	for i, b := range e.m_worldView {
		if b == nil || (b.m_ID == incomingBackup.m_ID) {
			e.m_worldView[i] = incomingBackup

			return
		}
	}
}

func (e *Elevator) TryUpdateWorldView(incomingBackup *Backup) bool {

	for _, b := range e.m_worldView {
		if b != nil && b.m_ID == incomingBackup.m_ID {
			return incomingBackup.m_version > b.m_version || !b.m_connectedToNetwork
		}
	}
	return true
}

func (e *Elevator) UpdateGlobalRequests(incomingBackup *Backup) {
	if incomingBackup.m_isMaster {
		e.m_globalRequests = incomingBackup.m_globalRequests
	}
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
func (e *Elevator) ShouldRedistributeOrders(incomingBackup *Backup) bool {
	//SHould redistribute if new backup changes obstruction status, or if we lose connection or if we gain connection, or if we change motorstopstatus
    for _, b := range e.m_worldView {
		if b != nil && b.m_ID == incomingBackup.m_ID {
			return (b.m_isObstructed != incomingBackup.m_isObstructed || b.GetHasMotorstop() != incomingBackup.GetHasMotorstop())
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

	e.m_floor = b.m_floor
	e.m_direction = b.m_direction
	//Restore global requests?? How to do this

	e.SetCabRequest(b.m_cabRequests)

	e.restoreMyBackup(b)

}


func (e *Elevator) ClearDisconnectedNodeQueue(){
	for _, b := range e.m_worldView {
		if b != nil && !b.m_connectedToNetwork {
			for f := 0; f < config.N_FLOORS; f++ {
				for btn := 0; btn < config.N_BUTTONS-1; btn++ {
					//b.m_requests[f][btn] = false
					fmt.Println("Fix me")
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

func (e *Elevator) GetLocalRequestAtFloor(floor int, btn driver.ButtonType) bool {
	return e.m_myRequests[floor][btn]
}

func (e *Elevator) GetLocalRequests() [config.N_FLOORS][config.N_BUTTONS]bool {
	return e.m_myRequests
}

func (e *Elevator) GetGlobalRequests() [config.N_FLOORS][config.N_BUTTONS-1]bool {
	return e.m_globalRequests
}

func (e *Elevator) SetGlobalRequests(globalRequests [config.N_FLOORS][config.N_BUTTONS-1]bool) {
	e.m_globalRequests = globalRequests
}

func (e *Elevator) GetCabRequests() [config.N_FLOORS]bool {
	var cabRequests [config.N_FLOORS]bool

	for floor := 0; floor < config.N_FLOORS; floor++ {
		cabRequests[floor] = e.m_myRequests[floor][driver.BT_Cab]
	}

	return cabRequests
}

func (e *Elevator) SetCabRequest(cabRequests [config.N_FLOORS]bool) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		e.m_myRequests[floor][driver.BT_Cab] = cabRequests[floor]
	}
}

func (e *Elevator) SetLocalRequest(floor int, btn driver.ButtonType, active bool) {
	e.m_myRequests[floor][btn] = active
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
