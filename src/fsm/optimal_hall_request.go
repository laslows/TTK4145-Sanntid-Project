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

	// simulate until done
	for {
		sort.Slice(states, func(i, j int) bool {
			return states[i].Time < states[j].Time
		})

		done := true
		if anyUnassigned(reqs) {
			done = false
		}
		if unvisitedAreImmediatelyAssignable(reqs, states) {
			assignImmediate(reqs, states)
			done = true
		}
		if done {
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

		reqs := elev.GetRequests() // [N_FLOORS][N_BUTTONS]bool

		for f := 0; f < numFloors; f++ {
			grid[f] = make([]bool, width)
			if config.INCLUDE_CAB {
				grid[f][2] = reqs[f][driver.BT_Cab]
			}
		}

		result[id] = grid
	}

	for f := 0; f < numFloors; f++ {
		for c := 0; c < 2; c++ {
			if reqs[f][c].m_active && reqs[f][c].m_assignedTo != "" {
				id := reqs[f][c].m_assignedTo
				result[id][f][c] = true
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

// Produces a [][]bool grid of hall requests matching fn(req).
func filterReq(reqs [][2]Req, fn func(Req) bool) [config.N_FLOORS][2]bool {
	var out [config.N_FLOORS][2]bool
	for f := 0; f < config.N_FLOORS && f < len(reqs); f++ {
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
			// Small tie-breaker so sort is deterministic
			Time: time.Duration(i) * time.Microsecond,
		})
	}
	return out
}

func performInitialMove(s *State, reqs [][2]Req) {
	doIdle := func() {
		for c := 0; c < 2; c++ {
			if reqs[s.State.GetFloor()][c].m_active {
				reqs[s.State.GetFloor()][c].m_assignedTo = s.ID
				s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond
			}
		}
	}

	switch s.State.GetBehaviour() {
	case elevator.DoorOpen:
		s.Time += time.Duration(config.DOOR_OPEN_DURATION) / 2 * time.Millisecond
		doIdle()
	case elevator.Idle:
		doIdle()
	case elevator.Moving:
		s.State.SetFloor(s.State.GetFloor() + int(s.State.GetDirection()))
		s.Time += time.Duration(config.TRAVEL_DURATION) / 2 * time.Millisecond
	}
}

func performSingleMove(s *State, reqs [][2]Req) { //TODO: Fix, fix withRequests
	e := elevator.WithRequests(s.State, filterReq(reqs, isUnassigned))

	onClear := func(c elevator.Button) {
		switch c {
		case elevator.HallUp, elevator.HallDown:
			reqs[s.State.GetFloor()][int(c)].m_assignedTo = s.ID
		case elevator.Cab:
			s.State.SetRequest(s.State.GetFloor(), driver.BT_Cab, false)
		}
	}

	switch s.State.GetBehaviour() {
	case elevator.Moving:
		if ShouldStop(e) {
			s.State.SetBehaviour(elevator.DoorOpen)
			s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond
			s.State = ClearAtCurrentFloor(s.State, onClear)
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
			s.State = ClearAtCurrentFloor(s.State, onClear)

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

func unvisitedAreImmediatelyAssignable(reqs [][2]Req, states []State) bool {
	// no remaining cab requests
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
	// If there’s an unassigned hall request at a floor with an elevator and no cab requests,
	// assign it to an elevator present at that floor.
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

				// Require no cab requests for this elevator
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
				s.Time += time.Duration(config.DOOR_OPEN_DURATION) * time.Millisecond

				// Stop after assigning this request to one elevator
				break
			}
		}
	}
}
