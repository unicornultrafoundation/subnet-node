package ipmanager

import "context"

type IPManagerStateType string

const (
	IPManagerStateTypeStatic  IPManagerStateType = "static"
	IPManagerStateTypeDynamic IPManagerStateType = "dynamic"
)

type IPManagerState interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	NextState()
	GetType() IPManagerStateType
}
