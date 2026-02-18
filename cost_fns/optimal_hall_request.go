package costfns

/* func OptimalHallRequests(hallReqs [][2]bool, elevatorStates map[string]LocalElevatorState) map[string][][]bool {
	in {
		auto numFloors = len(hallReqs);
		assert (len(elevatorStates) > 0, "No Elevator states provided");
		assert

} */

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

func FilterReq(Req[2][] reqs){
	// TODO: Implement
}

func ToReqs(bool[2][] hallReqs) Req[2][] { //TODO: FIX
	reqs := make(Req[2][], len(hallReqs))
	for floor := 0; floor < len(hallReqs); floor++ {
		reqs[floor][0] = Req{active: hallReqs[floor][0], assignedTo: string.init}
		reqs[floor][1] = Req{active: hallReqs[floor][1], assignedTo: string.init}
	}
	return reqs;
}

func WithReqs(s State, reqs Req[2][]) ElevatorState {
	return s.state.WithRequests(reqs.FilterReq!(fn));
}

func AnyUnassigned(Req[2][] reqs) bool {
	return reqs
		.filterReq!(IsUnassigned)
		// TODO: add map
}

func InitialStates(states LocalElevatorState[string]) State[] {
	return zip(states.keys, states.values, fn(s) State{id: s.key, state: s.value, time: 0});
	// TODO: fix
}

func PerformInitialMove(ref State s, ref Req[2][] reqs) State {
}

func PerformSingleMove(ref State s, ref Req[2][] reqs) State {
}

func UnvisitedAreImmediatelyAssignable(Req[2][] reqs, State[] states) bool {
}

func AssignImmediate(ref Req[2][] reqs, ref State[] states) {
}
//tests under