//go:build windows && !amd64

package vpad

import "errors"

// NewUinput is available on Windows amd64, which is supported by ViGEmBus.
func NewUinput(_ string) (Pad, error) {
	return nil, errors.New("Windows virtual Xbox controllers require an amd64 build and ViGEmBus")
}
