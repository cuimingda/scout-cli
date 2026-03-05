package cmd

import (
	"fmt"
	"net/url"
)

func validateConnectionURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		return fmt.Errorf("missing protocol")
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	return nil
}

func validateConnectionURLs(args []string) []error {
	var errs []error
	for _, raw := range args {
		if err := validateConnectionURL(raw); err != nil {
			errs = append(errs, fmt.Errorf("invalid URL %q: %w", raw, err))
		}
	}
	return errs
}
