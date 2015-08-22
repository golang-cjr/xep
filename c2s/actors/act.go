package actors

import (
	"sync"
	"xep/c2s/stream"
)

type Continue interface {
	Do(func(stream.Stream) error, ...func(error)) Continue
	Run()
}

type step struct {
	s   stream.Stream
	act func(stream.Stream) error
	err func(error)
}

type cont struct {
	s     stream.Stream
	steps []step
}

func With(s stream.Stream) Continue {
	return &cont{s: s}
}

func (c *cont) Do(fn func(stream.Stream) error, err ...func(error)) Continue {
	s := step{s: c.s}
	s.act = fn
	if len(err) > 0 {
		s.err = err[0]
	}
	c.steps = append(c.steps, s)
	return c
}

func (c *cont) Run() {
	var next *cont
	var this *step
	if len(c.steps) > 0 {
		this = &c.steps[0]
	}
	if len(c.steps) > 1 {
		next = &cont{s: c.s}
		next.steps = c.steps[1:]
	}
	if this != nil {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func(this *step, next *cont) {
			if err := this.act(this.s); err == nil {
				if next != nil {
					next.Run()
				}
				wg.Done()
			} else if this.err != nil {
				this.err(err)
				wg.Done()
			} else {
				panic(err)
			}
		}(this, next)
		wg.Wait()
	}
}
