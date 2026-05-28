package leaderfollower

import "errors"

var (
	ErrDiscoveryFailed = errors.New("leaderfollower: peer discovery failed")
	ErrPeerCallFailed  = errors.New("leaderfollower: peer call failed")
)
