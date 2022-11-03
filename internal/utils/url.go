package utils

import (
	"fmt"
	"net/url"
)

func CheckURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q for URL", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("missing host for URL")
	}
	return nil
}
