package elevator

import (
	"Sanntid/src/config"
	"encoding/json"
)

type Backup struct {
	m_ID                 int
	m_floor              int
	m_direction          Direction
	m_requests           [config.N_FLOORS][config.N_BUTTONS]bool
	m_isMaster           bool
	m_behaviour          ElevatorBehaviour
	m_version            int
	m_connectedToNetwork bool
}

// Må lage egendefinert json-Marshaller og unmarshaller fordi json ikke klarer å håndtere egendefinert type
// (Direction) og fordi vi ikke kan eksportere medlemsvariabler i Backup

func (b *Backup) MarshalJSON() ([]byte, error) {
	type BackupJSON struct {
		ID                 int
		Floor              int
		Direction          int
		Requests           [config.N_FLOORS][config.N_BUTTONS]bool
		IsMaster           bool
		Version            int
		ConnectedToNetwork bool
		Behaviour          int
	}

	return json.Marshal(&BackupJSON{
		ID:                 b.m_ID,
		Floor:              b.m_floor,
		Direction:          int(b.m_direction),
		Requests:           b.m_requests,
		IsMaster:           b.m_isMaster,
		Version:            b.m_version,
		ConnectedToNetwork: b.m_connectedToNetwork,
		Behaviour:          int(b.m_behaviour),
	})
}

// Egendefinert unmarshaler
func (b *Backup) UnmarshalJSON(data []byte) error {
	type BackupJSON struct {
		ID                 int
		Floor              int
		Direction          int
		Requests           [config.N_FLOORS][config.N_BUTTONS]bool
		IsMaster           bool
		Version            int
		Behaviour          int
		ConnectedToNetwork bool
	}

	var backupJSON BackupJSON
	err := json.Unmarshal(data, &backupJSON)
	if err != nil {
		return err
	}

	b.m_ID = backupJSON.ID
	b.m_floor = backupJSON.Floor
	b.m_direction = Direction(backupJSON.Direction)
	b.m_requests = backupJSON.Requests
	b.m_isMaster = backupJSON.IsMaster
	b.m_version = backupJSON.Version
	b.m_connectedToNetwork = backupJSON.ConnectedToNetwork
	b.m_behaviour = ElevatorBehaviour(backupJSON.Behaviour)

	return nil
}

// This is ugly..
// We will tidy up later :D
func (e *Elevator) UpdateMyBackup() {
	e.m_myBackup.m_version++
	e.m_myBackup.m_isMaster = e.m_isMaster
	e.m_myBackup.m_direction = e.m_direction
	e.m_myBackup.m_floor = e.m_floor
	e.m_myBackup.m_requests = e.m_requests
	e.m_myBackup.m_connectedToNetwork = true //Should always be true for master
	e.m_myBackup.m_behaviour = e.m_behaviour

	e.UpdateWorldView(e.m_myBackup)
}

func (b *Backup) GetID() int {
	return b.m_ID
}

func (b *Backup) GetBehaviour() ElevatorBehaviour {
	return b.m_behaviour
}

func (b *Backup) GetDirection() Direction {
	return b.m_direction
}

func (b *Backup) GetFloor() int {
	return b.m_floor
}

func (b *Backup) GetRequests() [config.N_FLOORS][config.N_BUTTONS]bool {
	return b.m_requests
}

func (b *Backup) GetConnectedToNetwork() bool {
	return b.m_connectedToNetwork
}

/*
func NewBackup(IP string) *Backup {
	return &Backup{
		m_IP: IP,
	}
}

func (b *Backup) GetIP() string {
	return b.m_IP
}
*/
