package closer

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
)

type closeFn struct {
	name string
	fn   func(context.Context) error
}

type closer struct {
	mu    sync.Mutex
	once  sync.Once
	funcs []closeFn
}

// Global closer.
var globalCloser = &closer{}

func Add(name string, fn func(context.Context) error) {
	globalCloser.add(name, fn)
}

func (c *closer) add(name string, fn func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, closeFn{name: name, fn: fn})
}

func CloseAll(ctx context.Context) error {
	return globalCloser.closeAll(ctx)
}

func (c *closer) closeAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			return
		}

		log := logger.Get()

		log.Info("closer: started closing resources", zap.Int("count", len(funcs)))

		var errs []error

		//	LIFO
		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]

			log.Info("closer: closing resource", zap.String("name", f.name))

			if err := f.fn(ctx); err != nil {
				log.Error("closer: failed to close resource",
					zap.String("name", f.name), zap.Error(err),
				)
				errs = append(errs, err)
			} else {
				log.Info("closer: resource closed", zap.String("name", f.name))
			}
		}

		result = errors.Join(errs...)
	})

	return result
}
