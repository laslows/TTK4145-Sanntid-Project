package orders

import (
	"encoding/json"
)

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

func (o *Order) MarshalJSON() ([]byte, error) {
	type OrderJSON struct {
		Floor     int
		OrderType int
	}

	return json.Marshal(&OrderJSON{
		Floor:     o.m_floor,
		OrderType: int(o.m_orderType),
	})
}

func (o *Order) UnmarshalJSON(data []byte) error {
	type OrderJSON struct {
		Floor     int
		OrderType int
	}

	var orderJSON OrderJSON
	err := json.Unmarshal(data, &orderJSON)
	if err != nil {
		return err
	}

	o.m_floor = orderJSON.Floor
	o.m_orderType = OrderType(orderJSON.OrderType)
	return nil
}
