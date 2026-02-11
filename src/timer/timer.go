package timer

import "time"

type Timer struct {
    m_endTime time.Time
    m_timerActive  bool
}

//Constructor. Lager en timer-struct og returnerer en peker til denne
func New() *Timer {
    return &Timer{}
}

func (t *Timer) Start(duration time.Duration) {
    t.endTime = time.Now().Add(duration)
    t.active = true
}

func (t *Timer) Stop() {
    t.active = false
}

func (t *Timer) TimedOut() bool {
    return t.active && time.Now().After(t.endTime)
}