package config

type ClearVariant int

const (
	CV_All ClearVariant = iota
	CV_InDirn
)

var clearRequestVariant = CV_InDirn
var doorOpenDuration float64 = 3.0
var inputPollRate int = 25

const N_FLOORS int = 4
const N_BUTTONS int = 3
