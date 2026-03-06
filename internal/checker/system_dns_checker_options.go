package checker

type SystemDNSFunc func() ([]string, error)

type SystemDNSCheckerOptions struct {
	SystemDNS SystemDNSFunc
}
