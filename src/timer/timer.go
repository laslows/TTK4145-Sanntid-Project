package timer

import (
	"time"
)

type Timer struct {
	m_endTime     time.Time
	m_timerActive bool
}

func New() *Timer {
	return &Timer{}
}

func (t *Timer) Start(duration time.Duration) {
	t.m_endTime = time.Now().Add(duration)
	t.m_timerActive = true
}

func (t *Timer) Stop() {
	t.m_timerActive = false
}

func (t *Timer) TimedOut() bool {
	return t.m_timerActive && time.Now().After(t.m_endTime)
}
