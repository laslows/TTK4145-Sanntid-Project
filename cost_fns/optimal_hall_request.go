package costfns

import (
	"runtime/debug"
	"time"
)


func OptimalHallRequests(hallReqs [][2]bool, elevatorStates [string]LocalElevatorState) [][][string]bool {
	in {
		auto numFloors := len(hallReqs);
		assert (len(elevatorStates) > 0, "No Elevator states provided");
		assert
		//TODO: implement
	}

	do {
		reqs := hallReqs.toReq;
		states := InitialStates(elevatorStates);

		for _, s := range states {
			performInitialMove(s, reqs);
		}

		while true {
			states.sort!("a.time < b.time")();

			//debug

			type done bool = true
			if reqs.anyUnassigned {
				done = false
			}
			if univistedAreImmediatelyAssignable(reqs, states) {
				//debug
				assignImmediate(reqs, states);
				done = true
			}

			if done {
				break;
			}

			//debug
		}
		type result [][][string]bool;

		if includeCab {}
		for _, id := range elevatorStates {
			result[id] = new bool[][](len(hallReqs), 3)
			for f := 0; f < len(hallReqs); f++ {
				results[id][f][2] = elevatorStates[id].cabRequests[f]
			}
		} 
	} else {
		for _, id := range elevatorStates {
			result[id] = new bool[][](len(hallReqs), 2)
		}
	}

	for f := 0; f < len(hallReqs); f++ {
		for c := 0; c < 2; c++ {
			if reqs[f][c].Active {
				result[reqs[f][c].AssignedTo][f][c] = true;
			}
		}
	}

	// debugs

	return result;
}

type Req struct	{
	bool active;
	string assignedTo;
}

type State struct {
	string id;
	LocalElevatorState state;
	Duration time;
}

func IsUnassigned(Req r) bool {
	return r.active && r.assignedTo == string.init;
}

func FilterReq(reqs [2][]Req, fn func(Req bool) [][2]bool) {
	out := make([][2]bool, len(reqs))
for f:= range reqs {
		out[f][0] = fn(reqs[f][0]);
		out[f][1] = fn(reqs[f][1]);
	}
	return out
}

func ToReqs(hallReqs [][2]bool) [][2]Req { //TODO: FIX
	reqs := make([][2]Req, len(hallReqs))
	for f := range hallReqs {
		for b := 0; b < 2; b++ {
			reqs[f][b] = Req{
				active: hallReqs[f][b],
				assignedTo: "",
			}
		}
	}
	return reqs;
}

func WithReqs(s State, reqs Req[2][]) ElevatorState {
	return s.state.WithRequests(reqs.FilterReq!(fn));
}

func AnyUnassigned(Req[2][] reqs) bool {
	for _, pair := range reqs {
		if isUnnassigned(pair[0]) || isUnassigned(pair[1]) {
			return true;
		}
	}
	return false;
}

func InitialStates(states map[string]LocalElevatorState) []State {
	keys := make([]string, 0, len(states))
	for k := range states {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]State, 0, len(states))
	for i, k := range keys {
		v := states[k]
		out = append(out, State{
			ID: k,
			Local: v,
			Age: time.Duration(i)*time.Microsecond

		})
	}
	return out;
}

func PerformInitialMove(s *State, reqs [][2]Req) {
	debug(optimalHallRequests, "Performing initial move for state %s with reqs %v", s.ID, reqs)

	doIdle := func() {
		for c := 0; c < 2; c++ {
			if reqs[s.State.Floor][c].Active {
				reqs[s.State.Floor][c].AssignedTo = s.ID
				s.Time += DoorOpenDuration * time.Millisecond

				fmt.Printf(" '%s' taking req %d at floor %d\n", s.ID, c, s.State.Floor)
			}
		}
	}

	switch s.State.Behaviour {
	case DoorOpen:
		debug(optimalHallRequests, "Closing door for %s at floor %d", s.ID, s.State.Floor)
		s.Time += (DoorOpenDuration / 2) * time.Millisecond
		doIdle()

	case Idle:
		doIdle()

	case Moving:
		s.State.Floor += int(s.State.Direction) // TODO: fix types
		s.Time += (TravelDuration / 2) * time.Millisecond
		debug(optimalHallRequests, "%s moving to floor %d", s.ID, s.State.Floor)
	}
}


func PerformSingleMove(ref State s, ref Req[2][] reqs) State {
	// auto e = s.withReqs!(isUnassigned)(reqs);
	//debug(optimal_hall_requests) writefln("s",e);

	auto onClearRequest = (CallType c){
		switch c CallType {
		case HallUp, HallDown:
			reqs[s.state.floor][c].AssignedTo = s.ID;
			break;
		case Cab:
			s.state.CabRequests[s.state.floor] = false;
		}
	}

	switch ElevatorBehaviour s.state.Behaviour {
	case Moving:
		if e.ShouldStop{
			s.state.behaviour = DoorOpen;
			s.Time += DoorOpenDuration * time.Millisecond
			e.clearReqsAtFloor(onClearRequest);
			debug(optimalHallRequests, "%s stopping at floor %d", s.ID, s.state.floor)
		} else {
			s.state.floor += s.state.direction
			s.time += TravelDuration * time.Millisecond
			debug(optimalHallRequests, "%s moving to floor %d", s.ID, s.state.floor)
		}
		break;
	case Idle, DoorOpen:
		s.state.direction = e.ChooseDirection
		if s.state.direction == Dirn.stop {
			if e.AnyRequestsAtFloor {
				e.ClearReqsAtFloor(onClearRequest);
				s.time += DoorOpenDuration * time.Millisecond
				s.state.behaviour = DoorOpen
				debug(optimalHallRequests, "%s opening door at floor %d", s.ID, s.state.floor)
		} else {
			s.state.behaviour = Moving
			s.state.floor += s.state.direction
			s.time += TravelDuration * time.Millisecond
			debug(optimalHallRequests, "%s starting to move %s from floor %d", s.ID, s.state.direction, s.state.floor)
		}
		break;
	}
	debug(optimal_hall_requests) writefln("s",s); //TODO: fix lmao
}

func UnvisitedAreImmediatelyAssignable(Req[2][] reqs, State[] states) bool {
	if states.map!(a => a.state.CabRequests.any).any {
		return false;
	}
	for _, reqsAtFloor := range reqs {
			if reqsAtFloor[0].Active && isUnassigned(reqsAtFloor[0]) {
				return true;
			}
			if reqsAtFloor[1].Active && isUnassigned(reqsAtFloor[1]) {
				return true;
			}
	}
	for c, req:= range reqsAtFloor {
		if req.IsUnassigned {
			if states.filter!(s => s.state.CabRequests[c]).any {
				return false;
			}
		}
	}
	return true;
}

func AssignImmediate(ref Req[2][] reqs, ref State[] states) {
	for f, reqsAtFloor := range reqs {
		for c, req := raange reqsAtFloor {
			for _, s:= range states {
				if req.IsUnassigned {
					if s.state.floor == f && !s.state.cabRequests.any {
						req.AssignedTo = s.ID
						s.time += DoorOpenDuration * time.Millisecond
					}
				}
			}
		}
}
//tests under