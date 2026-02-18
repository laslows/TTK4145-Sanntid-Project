package orders

//TODO: FIx naming conventions

type OrderType int

const (
	HALL_UP OrderType = iota
	HALL_DOWN
	CAB
)

type Order struct {
	m_floor     int
	m_orderType OrderType
}

// Continue on this
func New(floor int, orderType OrderType) Order {
	return Order{
		m_floor:     floor,
		m_orderType: orderType,
	}
}

func (o Order) GetFloor() int {
	return o.m_floor
}

func (o Order) GetOrderType() OrderType {
	return o.m_orderType
}
