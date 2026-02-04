package fsm

//TODO: Fix naming conventions

type State int

const (
	Idle State = iota
	AtFloor
	Moving
	MotorStop
)

func Fsm() {
	//Tilstandsmaskin
}