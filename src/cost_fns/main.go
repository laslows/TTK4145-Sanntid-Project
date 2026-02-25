package costfns

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	config "./config"
)

type Input struct {
	HallRequests [][]bool                      `json:"hallRequests"`
	States       map[string]LocalElevatorState `json:"states"`
}

func main() {

	var inputStr string
	var clearReq string

	flag.IntVar(&config.DoorOpenDuration, "doorOpenDuration", config.DoorOpenDuration, "Duration the door is open in milliseconds")
	flag.IntVar(&config.TravelDuration, "travelDuration", config.TravelDuration, "Duration it takes to travel one floor in milliseconds")
	flag.BoolVar(&config.IncludeCab, "includeCab", config.IncludeCab, "Whether to include cab requests in the cost function")
	flag.StringVar(&clearReq, "clearReq", "inDirn", "Whether to clear all requests or only those in the direction of travel (all/inDirn)")
	flag.StringVar(&inputStr, "input", "", "Input string in the format '[[hallReqs]], {elevatorStates}'")

	flag.Parse()

	switch clearReq {
	case "All", "all":
		config.SetClearRequestType(config.All)
	case "InDirn", "inDirn":
		config.SetClearRequestType(config.InDirn)
	default:
		fmt.Fprintf(os.Stderr, "Invalid clearReq value. Use 'all' or 'inDirn'.\n")
		os.Exit(2)
	}
	//TODO: get DoorOpenDuration etc?

	if inputStr == "" {
		b, err := io.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		inputStr = string(b)
	}

	var in Input
	if err := json.Unmarshal([]byte(inputStr), &in); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input: %v\n", err)
		os.Exit(1)
	}

	//TODO: Fix main
	//OptimalHallRequests := OptimalHallRequests(in.HallRequests, in.States)
	hall := in.HallRequests
	out := config.OptimalHallRequests(hall, in.States)

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding output: %v\n", err)
		os.Exit(1)
	}

}

/* unitest {s
	//TODO: find corret go library and write some tests
} */
