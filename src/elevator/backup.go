package elevator

import (
	"Sanntid/src/config"
	"encoding/json"
)

type Backup struct {
	m_IP                 string
	m_port               string
	m_floor              int
	m_direction          Direction
	m_requests           [config.N_FLOORS][config.N_BUTTONS]bool
	m_isMaster           bool
	m_version            int
	m_connectedToNetwork bool
}

// Må lage egendefinert json-Marshaller og unmarshaller fordi json ikke klarer å håndtere egendefinert type
// (Direction) og fordi vi ikke kan eksportere medlemsvariabler i Backup

func (b *Backup) MarshalJSON() ([]byte, error) {
	type BackupJSON struct {
		IP                 string
		Port               string
		Floor              int
		Direction          int
		Requests           [config.N_FLOORS][config.N_BUTTONS]bool
		IsMaster           bool
		Version            int
		ConnectedToNetwork bool
	}

	return json.Marshal(&BackupJSON{
		IP:                 b.m_IP,
		Port:               b.m_port,
		Floor:              b.m_floor,
		Direction:          int(b.m_direction),
		Requests:           b.m_requests,
		IsMaster:           b.m_isMaster,
		Version:            b.m_version,
		ConnectedToNetwork: b.m_connectedToNetwork,
	})
}

// Egendefinert unmarshaler
func (b *Backup) UnmarshalJSON(data []byte) error {
	type BackupJSON struct {
		IP                 string
		Port               string
		Floor              int
		Direction          int
		Requests           [config.N_FLOORS][config.N_BUTTONS]bool
		IsMaster           bool
		Version            int
		ConnectedToNetwork bool
	}

	var backupJSON BackupJSON
	err := json.Unmarshal(data, &backupJSON)
	if err != nil {
		return err
	}

	b.m_IP = backupJSON.IP
	b.m_port = backupJSON.Port
	b.m_floor = backupJSON.Floor
	b.m_direction = Direction(backupJSON.Direction)
	b.m_requests = backupJSON.Requests
	b.m_isMaster = backupJSON.IsMaster
	b.m_version = backupJSON.Version
	b.m_connectedToNetwork = backupJSON.ConnectedToNetwork

	return nil
}

// This is ugly..
// We will tidy up later :D
func (e *Elevator) UpdateOwnBackup() {
	e.m_myBackup.m_version++
	e.m_myBackup.m_isMaster = e.m_isMaster
	e.m_myBackup.m_direction = e.m_direction
	e.m_myBackup.m_floor = e.m_floor
	e.m_myBackup.m_requests = e.m_requests
	e.m_myBackup.m_connectedToNetwork = true //Should always be true..

	e.UpdateWorldView(e.m_myBackup)
}

func (b *Backup) GetPort() string {
	return b.m_port
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
