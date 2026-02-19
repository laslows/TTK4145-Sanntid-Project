package elevator

import (
	"Sanntid/src/config"
)

type Backup struct {
	m_IP        string
	m_port	    string
	m_floor     int
	m_direction Direction
	m_requests  [config.N_FLOORS][config.N_BUTTONS]bool
	m_isMaster  bool
	m_connectedToNetwork bool
}

//This is ugly..
//We will tidy up later :D
func (e *Elevator) UpdateOwnBackup() {
	e.m_myBackup.m_isMaster = e.m_isMaster
	e.m_myBackup.m_direction = e.m_direction
	e.m_myBackup.m_floor = e.m_floor
	e.m_myBackup.m_requests = e.m_requests
	e.m_myBackup.m_connectedToNetwork = true //Should always be true..

	e.UpdateWorldView(e.m_myBackup)
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
