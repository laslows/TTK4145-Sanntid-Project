package costfns

type Input struct {
	bool               [][]HallRequests
	LocalElevatorState [string]States
}

func main(string []args) void {

	string input_str;

	//TODO: get DoorOpenDuration etc?

	if input_str == string.init {
		input_str = readln;
		input_str = input_str[1:len(input_str)-1]; //removes trailing
	}

	Input i = inputstr.Input;

	//TODO: Fix main

}

unitest {
	//TODO: find corret go library and write some tests
}