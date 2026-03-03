package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"sort"
	"time"
)

type Req struct {
	m_active     bool
	m_assignedTo string
}

type State struct {
	ID    string
	State elevator.Elevator
	Time  time.Duration
}

func OptimalHallRequests(hallReqs [][2]bool, elevatorStates map[string]elevator.Elevator) map[string][][]bool {
	validateInputs(hallReqs, elevatorStates)

	reqs := toReqs(hallReqs)
	states := initialStates(elevatorStates)

	// initial moves
	for i := range states {
		performInitialMove(&states[i], reqs)
	}

	// simulate until all hall reqs are assigned
	for {
		sort.Slice(states, func(i, j int) bool {
			return states[i].Time < states[j].Time
		})

		if !anyUnassigned(reqs) {
			break
		}

		if unvisitedAreImmediatelyAssignable(reqs, states) {
			assignImmediate(reqs, states)
			break
		}

		performSingleMove(&states[0], reqs)
	}

	numFloors := len(hallReqs)
	width := 2
	if config.INCLUDE_CAB {
		width = 3
	}

	result := make(map[string][][]bool, len(elevatorStates))
	for id, elev := range elevatorStates {
		grid := make([][]bool, numFloors)

		elevReqs := elev.GetRequests() // [N_FLOORS][N_BUTTONS]bool
		for f := 0; f < numFloors; f++ {
			grid[f] = make([]bool, width)
			if config.INCLUDE_CAB {
				grid[f][2] = elevReqs[f][driver.BT_Cab]
			}
		}
		result[id] = grid
	}

	// Apply hall assignments
	for f := 0; f < numFloors; f++ {
		for c := 0; c < 2; c++ {
			if reqs[f][c].m_active && reqs[f][c].m_assignedTo != "" {
				id := reqs[f][c].m_assignedTo
				if _, ok := result[id]; ok {
					result[id][f][c] = true
				}
			}
		}
	}

	return result
}

func validateInputs(hallReqs [][2]bool, elevatorStates map[string]elevator.Elevator) {
	numFloors := len(hallReqs)

	if len(elevatorStates) == 0 {
		panic("no elevator states provided")
	}

	isInBounds := func(f int) bool { return f >= 0 && f < numFloors }

	for _, st := range elevatorStates {
		if len(st.GetRequests()) != numFloors {
			panic("hall and cab requests do not all have the same length")
		}
		if !isInBounds(st.GetFloor()) {
			panic("some elevator is at an invalid floor")
		}
		if st.GetBehaviour() == elevator.Moving && !isInBounds(st.GetFloor()+int(st.GetDirection())) {
			panic("some elevator is moving away from an end floor")
		}
	}
}

func isUnassigned(r Req) bool {
	return r.m_active && r.m_assignedTo == ""
}

func anyUnassigned(reqs [][2]Req) bool {
	for _, pair := range reqs {
		if isUnassigned(pair[0]) || isUnassigned(pair[1]) {
			return true
		}
	}
	return false
}

func filterReq(reqs [][2]Req, fn func(Req) bool) [config.N_FLOORS][2]bool {
	var out [config.N_FLOORS][2]bool
	maxF := len(reqs)
	if maxF > config.N_FLOORS {
		maxF = config.N_FLOORS
	}
	for f := 0; f < maxF; f++ {
		out[f][0] = fn(reqs[f][0])
		out[f][1] = fn(reqs[f][1])
	}
	return out
}

func toReqs(hallReqs [][2]bool) [][2]Req {
	reqs := make([][2]Req, len(hallReqs))
	for f := range hallReqs {
		for b := 0; b < 2; b++ {
			reqs[f][b] = Req{
				m_active:     hallReqs[f][b],
				m_assignedTo: "",
			}
		}
	}
	return reqs
}

func initialStates(states map[string]elevator.Elevator) []State {
	keys := make([]string, 0, len(states))
	for k := range states {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]State, 0, len(states))
	for i, k := range keys {
		v := states[k]
		out = append(out, State{
			ID:    k,
			State: v,
			Time:  time.Duration(i) * time.Microsecond,
		})
	}
	return out
}

func performInitialMove(s *State, reqs [][2]Req) {
	f := s.State.GetFloor()
	if f < 0 || f >= len(reqs) {
		return
	}

	assignHere := func() bool {
		assignedAny := false
		for c := 0; c < 2; c++ {
			if reqs[f][c].m_active && reqs[f][c].m_assignedTo == "" {
				reqs[f][c].m_assignedTo = s.ID
				assignedAny = true
			}
		}
		return assignedAny
	}

	switch s.State.GetBehaviour() {
	case elevator.DoorOpen:
		// Assume door is halfway through its open time; finish remaining half.
		s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond / 2
		// If we serve hall calls while door is already open, don't add extra door time.
		_ = assignHere()

	case elevator.Idle:
		// If idle and there are hall calls here, we open the door once to serve them.
		if assignHere() {
			s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond
		}

	case elevator.Moving:
		// Assuming it’s halfway to next floor...
		s.State.SetFloor(s.State.GetFloor() + int(s.State.GetDirection()))
		s.Time += time.Duration(config.TRAVEL_DURATION) * time.Millisecond / 2
	}
}

// Core simulation step for the earliest-available elevator.
func performSingleMove(s *State, reqs [][2]Req) {
	e := elevator.WithRequests(s.State, filterReq(reqs, isUnassigned))

	switch s.State.GetBehaviour() {

	case elevator.Moving:
		if ShouldStop(e) {
			s.State.SetBehaviour(elevator.DoorOpen)
			s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond
			s.State = clearAndRecordServedRequests(s.ID, s.State, reqs)
		} else {
			s.State.SetFloor(s.State.GetFloor() + int(s.State.GetDirection()))
			s.Time += time.Duration(config.TRAVEL_DURATION) * time.Millisecond
		}

	case elevator.Idle, elevator.DoorOpen:
		pair := ChooseDirection(e)
		s.State.SetDirection(pair.m_dirn)

		switch pair.m_behaviour {

		case elevator.DoorOpen:
			s.State.SetBehaviour(elevator.DoorOpen)
			s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond
			s.State = clearAndRecordServedRequests(s.ID, s.State, reqs)

		case elevator.Moving:
			s.State.SetBehaviour(elevator.Moving)
			s.State.SetFloor(s.State.GetFloor() + int(s.State.GetDirection()))
			s.Time += time.Duration(config.TRAVEL_DURATION) * time.Millisecond

		case elevator.Idle:
			s.State.SetBehaviour(elevator.Idle)

		default:
			s.State.SetBehaviour(pair.m_behaviour)
		}
	}
}

func clearAndRecordServedRequests(elevID string, e elevator.Elevator, reqs [][2]Req) elevator.Elevator {
	f := e.GetFloor()
	if f < 0 || f >= len(reqs) {
		return e
	}

	beforeUp := e.GetRequestAtFloor(f, driver.BT_HallUp)
	beforeDown := e.GetRequestAtFloor(f, driver.BT_HallDown)

	e2 := ClearAtCurrentFloor(e)

	afterUp := e2.GetRequestAtFloor(f, driver.BT_HallUp)
	afterDown := e2.GetRequestAtFloor(f, driver.BT_HallDown)

	// If request existed and got cleared, it was served now.
	if beforeUp && !afterUp && reqs[f][int(elevator.HallUp)].m_active && reqs[f][int(elevator.HallUp)].m_assignedTo == "" {
		reqs[f][int(elevator.HallUp)].m_assignedTo = elevID
	}
	if beforeDown && !afterDown && reqs[f][int(elevator.HallDown)].m_active && reqs[f][int(elevator.HallDown)].m_assignedTo == "" {
		reqs[f][int(elevator.HallDown)].m_assignedTo = elevID
	}

	return e2
}

func unvisitedAreImmediatelyAssignable(reqs [][2]Req, states []State) bool {
	// no remaining cab requests anywhere
	for _, s := range states {
		eReqs := s.State.GetRequests()
		for f := 0; f < config.N_FLOORS; f++ {
			if eReqs[f][driver.BT_Cab] {
				return false
			}
		}
	}

	for f := range reqs {
		// no floors with two active hall requests
		activeCount := 0
		if reqs[f][0].m_active {
			activeCount++
		}
		if reqs[f][1].m_active {
			activeCount++
		}
		if activeCount == 2 {
			return false
		}

		// all unassigned must be at floors with elevators
		for c := 0; c < 2; c++ {
			if isUnassigned(reqs[f][c]) {
				found := false
				for _, s := range states {
					if s.State.GetFloor() == f {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	return true
}

func assignImmediate(reqs [][2]Req, states []State) {

	doorAddedAtFloor := make(map[string]map[int]bool) // elevID -> floor -> bool

	markDoor := func(elevID string, floor int) {
		if _, ok := doorAddedAtFloor[elevID]; !ok {
			doorAddedAtFloor[elevID] = make(map[int]bool)
		}
		doorAddedAtFloor[elevID][floor] = true
	}

	doorAlready := func(elevID string, floor int) bool {
		if m, ok := doorAddedAtFloor[elevID]; ok {
			return m[floor]
		}
		return false
	}

	for f := range reqs {
		for c := 0; c < 2; c++ {
			if !isUnassigned(reqs[f][c]) {
				continue
			}

			for si := range states {
				s := &states[si]
				if s.State.GetFloor() != f {
					continue
				}

				hasCab := false
				eReqs := s.State.GetRequests()
				for floor := 0; floor < config.N_FLOORS; floor++ {
					if eReqs[floor][driver.BT_Cab] {
						hasCab = true
						break
					}
				}
				if hasCab {
					continue
				}

				reqs[f][c].m_assignedTo = s.ID

				if !doorAlready(s.ID, f) {
					s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond
					markDoor(s.ID, f)
				}

				break
			}
		}
	}
}
