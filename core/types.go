package core

import "errors"

func A() error {
	return errors.New("error")
}
