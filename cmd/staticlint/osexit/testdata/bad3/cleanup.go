package main

type cleanuper struct {
	cleanups []func()
}

func (c *cleanuper) Add(fns ...func()) {
	c.cleanups = append(c.cleanups, fns...)
}

func (c *cleanuper) Run() {
	defer func() {
		if len(c.cleanups) > 0 {
			c.Run()
		}
	}()

	for {
		var cleanup func()

		if len(c.cleanups) > 0 {
			last := len(c.cleanups) - 1
			cleanup = c.cleanups[last]
			c.cleanups = c.cleanups[:last]
		}

		if cleanup == nil {
			return
		}

		cleanup()
	}
}
