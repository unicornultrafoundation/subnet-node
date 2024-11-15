package routing

import (
	"fmt"
)

type ParamNeededError struct {
	ParamName  string
	RouterType RouterType
}

func NewParamNeededErr(param string, routing RouterType) error {
	return &ParamNeededError{
		ParamName:  param,
		RouterType: routing,
	}
}

func (e *ParamNeededError) Error() string {
	return fmt.Sprintf("configuration param '%v' is needed for %v delegated routing types", e.ParamName, e.RouterType)
}
