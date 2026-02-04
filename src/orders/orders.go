package orders

//TODO: FIx naming conventions

type OrderType int

const (
	HALL_UP OrderType = iota
	HALL_DOWN
	CAB
)

type Order struct {
	Floor int
	OrderType OrderType
}