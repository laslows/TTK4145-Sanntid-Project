package elevator

import (
	"Sanntid/src/config"
)

type Backup struct {
	m_IP        string
	m_floor     int
	m_direction Direction
	m_requests  [config.N_FLOORS][config.N_BUTTONS]bool
	m_isMaster  bool
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
