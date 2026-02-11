package timer

import (
	"fmt"
	"time"
)

type Timer struct {
	m_endTime     time.Time
	m_timerActive bool
}

// Constructor. Lager en timer-struct og returnerer en peker til denne
func New() *Timer {
	return &Timer{}
}

func (t *Timer) Start(duration time.Duration) {
	fmt.Println("Start timer")
	t.m_endTime = time.Now().Add(duration)
	t.m_timerActive = true
}

func (t *Timer) Stop() {
	fmt.Println("Stopped timer")
	t.m_timerActive = false
}

func (t *Timer) TimedOut() bool {
	return t.m_timerActive && time.Now().After(t.m_endTime)
}
