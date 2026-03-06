package checker

import "time"

type DialFunc func(network, address string, timeout time.Duration) error

type PortCheckerOptions struct {
	Dial    DialFunc
	Timeout time.Duration
}
