package frontend

import "GoInNuvola/core"

func NewFrontend() (core.Frontend, error) {
	return &RestFrontend{}, nil
}
