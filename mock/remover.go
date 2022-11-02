package mock

import "context"

type Remover struct {
	RemoveInvoked bool
	RemoveFn      func(string) error
}

func NewRemover() *Remover {
	return &Remover{RemoveFn: func(string) error { return nil }}
}

func (r *Remover) Remove(ctx context.Context, name string) error {
	r.RemoveInvoked = true
	return r.RemoveFn(name)
}
