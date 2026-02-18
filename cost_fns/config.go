package costfns

const (
	doorOpenDuration  int  = 3000;
	travelDuration  int    = 2500;
	includeCab bool         = false;
)

type ClearRequestType int

const (
	all ClearRequestType = iota
	inDirn
)

// TODO: Fix
ClearRequestType clearRequestType = ClearRequestType.inDirn;