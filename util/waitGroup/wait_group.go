package waitGroup

import (
	"sync"
)

type Wrapper struct {
	sync.WaitGroup
}

func (w *Wrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

func (w *Wrapper) WrapWithRecover(cb func(), recoverHandler func(r interface{})) {
	w.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if recoverHandler != nil {
					recoverHandler(r)
				}
			}
			w.Done()
		}()
		cb()
	}()
}
