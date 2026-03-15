package elevator

import (
	"Sanntid/src/driver"
)

func FloorSensor() int {
	return driver.GetFloor()
}

func RequestButton(floor int, btn driver.ButtonType) bool {
	return driver.GetButton(btn, floor)
}

func ObstructionSwitch() bool {
	return driver.GetObstruction()
}

func FloorIndicator(floor int) {
	driver.SetFloorIndicator(floor)
}

func RequestButtonLight(floor int, btn driver.ButtonType, on bool) {
	driver.SetButtonLamp(btn, floor, on)
}

func DoorOpenLight(on bool) {
	driver.SetDoorOpenLamp(on)
}

func MotorDirection(dir Direction) {
	driver.SetMotorDirection(driver.MotorDirection(dir))
}