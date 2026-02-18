package costfns

import (
	"fmt"
	"runtime/debug"
	"sort"
	"time"

	ea "./elevator_algorithm"
	es "./elevator_state"
)

// ------------------------------------------------------------
// Types
// ------------------------------------------------------------

type Req struct {
	Active     bool
	AssignedTo string
}

type State struct {
	ID    string
	State es.LocalElevatorState
	Time  time.Duration
}

// ------------------------------------------------------------
// Public entry
// ------------------------------------------------------------

func OptimalHallRequests(hallReqs [][2]bool, elevatorStates map[string]es.LocalElevatorState) map[string][][]bool {
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
	if IncludeCab {
		width = 3
	}

	result := make(map[string][][]bool, len(elevatorStates))
	for id := range elevatorStates {
		grid := make([][]bool, numFloors)
		for f := 0; f < numFloors; f++ {
			grid[f] = make([]bool, width)
			if IncludeCab {
				grid[f][2] = elevatorStates[id].CabRequests[f]
			}
		}
		result[id] = grid
	}

	for f := 0; f < numFloors; f++ {
		for c := 0; c < 2; c++ {
			if reqs[f][c].Active && reqs[f][c].AssignedTo != "" {
				id := reqs[f][c].AssignedTo
				result[id][f][c] = true
			}
		}
	}

	return result
}

// ------------------------------------------------------------
// Validation
// ------------------------------------------------------------

func validateInputs(hallReqs [][2]bool, elevatorStates map[string]es.LocalElevatorState) {
	numFloors := len(hallReqs)

	if len(elevatorStates) == 0 {
		panic("no elevator states provided")
	}

	isInBounds := func(f int) bool { return f >= 0 && f < numFloors }

	for _, st := range elevatorStates {
		if len(st.CabRequests) != numFloors {
			panic("hall and cab requests do not all have the same length")
		}
		if !isInBounds(st.Floor) {
			panic("some elevator is at an invalid floor")
		}
		if st.Behaviour == es.Moving && !isInBounds(st.Floor+int(st.Direction)) {
			panic("some elevator is moving away from an end floor")
		}
	}
}

// ------------------------------------------------------------
// Req helpers
// ------------------------------------------------------------

func isUnassigned(r Req) bool {
	return r.Active && r.AssignedTo == ""
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
func filterReq(reqs [][2]Req, fn func(Req) bool) [][2]bool {
	out := make([][2]bool, len(reqs))
	for f := range reqs {
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
				Active:     hallReqs[f][b],
				AssignedTo: "",
			}
		}
	}
	return reqs
}

// ------------------------------------------------------------
// State initialization
// ------------------------------------------------------------

func initialStates(states map[string]es.LocalElevatorState) []State {
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

// ------------------------------------------------------------
// Simulation steps
// ------------------------------------------------------------

func performInitialMove(s *State, reqs [][2]Req) {
	doIdle := func() {
		for c := 0; c < 2; c++ {
			if reqs[s.State.Floor][c].Active {
				reqs[s.State.Floor][c].AssignedTo = s.ID
				s.Time += time.Duration(DoorOpenDuration) * time.Millisecond
			}
		}
	}

	switch s.State.Behaviour {
	case es.DoorOpen:
		s.Time += time.Duration(DoorOpenDuration/2) * time.Millisecond
		doIdle()
	case es.Idle:
		doIdle()
	case es.Moving:
		s.State.Floor += int(s.State.Direction)
		s.Time += time.Duration(TravelDuration/2) * time.Millisecond
	}
}

func performSingleMove(s *State, reqs [][2]Req) {
	// Only consider unassigned hall reqs as "active" for decision-making
	e := ea.WithRequests(s.State, filterReq(reqs, isUnassigned))

	onClear := func(c es.CallType) {
		switch c {
		case es.HallUp, es.HallDown:
			reqs[s.State.Floor][int(c)].AssignedTo = s.ID
		case es.Cab:
			s.State.CabRequests[s.State.Floor] = false
		}
	}

	switch s.State.Behaviour {
	case es.Moving:
		if e.ShouldStop() {
			s.State.Behaviour = es.DoorOpen
			s.Time += time.Duration(DoorOpenDuration) * time.Millisecond
			e.ClearReqsAtFloor(onClear)
		} else {
			s.State.Floor += int(s.State.Direction)
			s.Time += time.Duration(TravelDuration) * time.Millisecond
		}

	case es.Idle, es.DoorOpen:
		s.State.Direction = e.ChooseDirection()

		if s.State.Direction == es.Stop {
			if e.AnyRequestsAtFloor() {
				e.ClearReqsAtFloor(onClear)
				s.Time += time.Duration(DoorOpenDuration) * time.Millisecond
				s.State.Behaviour = es.DoorOpen
			} else {
				s.State.Behaviour = es.Idle
			}
		} else {
			s.State.Behaviour = es.Moving
			s.State.Floor += int(s.State.Direction)
			s.Time += time.Duration(TravelDuration) * time.Millisecond
		}
	}
}

func unvisitedAreImmediatelyAssignable(reqs [][2]Req, states []State) bool {
	// no remaining cab requests
	for _, s := range states {
		for _, cr := range s.State.CabRequests {
			if cr {
				return false
			}
		}
	}

	for f := range reqs {
		// no floors with two active hall requests
		activeCount := 0
		if reqs[f][0].Active {
			activeCount++
		}
		if reqs[f][1].Active {
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
					if s.State.Floor == f {
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
	// If there’s an unassigned hall request at a floor *with an elevator* and there are no cab reqs,
	// just assign it to an elevator present at that floor.
	for f := range reqs {
		for c := 0; c < 2; c++ {
			if !isUnassigned(reqs[f][c]) {
				continue
			}

			for si := range states {
				s := &states[si]

				if s.State.Floor != f {
					continue
				}

				// require no cab requests (matches your earlier condition)
				hasCab := false
				for _, cr := range s.State.CabRequests {
					if cr {
						hasCab = true
						break
					}
				}
				if hasCab {
					continue
				}

				reqs[f][c].AssignedTo = s.ID
				s.Time += time.Duration(DoorOpenDuration) * time.Millisecond
			}
		}
	}
}

// ------------------------------------------------------------
// Optional debug helper (your original had debug(...) calls)
// ------------------------------------------------------------

func dbg(tag string, format string, args ...any) {
	_ = debug.Stack // keeps import if you want stack dumps later
	_ = tag
	fmt.Printf(format+"\n", args...)
}
