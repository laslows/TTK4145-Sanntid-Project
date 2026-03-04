package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
)

//TODO:fix name

type SystemState struct {
	HallRequests [][]bool
	States       map[string]ElevatorState
}

type ElevatorState struct {
	Behaviour   string
	Floor       int
	Direction   string
	CabRequests []bool
}

func createJSONDataForHallReqAlgorithm(e *elevator.Elevator) []byte {
	states := make(map[string]ElevatorState)
	hallRequests := make([][]bool, config.N_FLOORS)
	for i := range hallRequests {
		hallRequests[i] = make([]bool, config.N_BUTTONS-1)
	}
	worldView := e.GetWorldView()
	for _, backup := range worldView {
		if backup != nil { //TODO:motorstop and disconnected
			requests := backup.GetRequests()
			states[string(backup.GetID())] = ElevatorState{
				Behaviour:   elevator.BehaviourToString(backup.GetBehaviour()),
				Floor:       backup.GetFloor(),
				Direction:   elevator.DirectionToString(backup.GetDirection()),
				CabRequests: make([]bool, len(requests)),
			}
			for i, row := range requests {
				states[string(backup.GetID())].CabRequests[i] = row[len(row)-1]
			}
			for i := range hallRequests {
				for j := range hallRequests[i] {
					hallRequests[i][j] = hallRequests[i][j] || requests[i][j]
				}
			}
		}
	}
	system := SystemState{
		HallRequests: hallRequests,
		States:       states,
	}
	jsonData, err := json.Marshal(system)
	if err != nil {
		panic(err)
	}
	return jsonData
}

func runHallReqAlgorithm(data []byte) {
	input := string(data) // Convert data to string
	err := exec.Command("./hall_request_assigner", "--input", input).Run()
	if err != nil {
		fmt.Printf("Error running hall request algorithm: %v\n", err)
	}
}


