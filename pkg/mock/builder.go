package mock

import (
	"context"

	fn "knative.dev/func/pkg/functions"
)

type Builder struct {
	BuildInvoked bool
	BuildFn      func(fn.Function) error
}

func NewBuilder() *Builder {
	return &Builder{
		BuildFn: func(fn.Function) error { return nil },
	}
}

func (i *Builder) Build(ctx context.Context, f fn.Function, _ []fn.Platform) error {
	i.BuildInvoked = true
	return i.BuildFn(f)
}
