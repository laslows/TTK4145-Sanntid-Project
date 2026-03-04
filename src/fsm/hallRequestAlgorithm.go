package fsm

import(
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
)

func SuperOptimalOrderAssignmentAlgorithm() int {
	return rand.Intn(2)
}





func SmokeTestHallRequestAssigner() (string, error) {
	input := map[string]interface{}{
		"hallRequests": [][]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		"states": map[string]interface{}{
			"one": map[string]interface{}{
				"behaviour":   "moving",
				"floor":       2,
				"direction":   "up",
				"cabRequests": []bool{false, false, true, true},
			},
			"two": map[string]interface{}{
				"behaviour":   "idle",
				"floor":       0,
				"direction":   "stop",
				"cabRequests": []bool{false, false, false, false},
			},
		},
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal smoke test input: %w", err)
	}

	cmd := exec.Command(
		"./src/fsm/hall_request_assigner/hall_request_assigner",
		"--input", string(inputBytes),
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("hall_request_assigner failed: %w, output: %s", err, string(out))
	}

	return string(out), nil
}


