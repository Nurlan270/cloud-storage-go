package closer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
)

type closeFn struct {
	name string
	fn   func() error
}

type closer struct {
	mu    sync.Mutex
	once  sync.Once
	funcs []closeFn
}

// Global closer.
var globalCloser = &closer{}

func Add(name string, fn func() error) {
	globalCloser.add(name, fn)
}

func (c *closer) add(name string, fn func() error) {
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
			// Context of current resource; with timeout of 2 seconds
			// to avoid blocking the shutdown process
			rCtx, rCancel := context.WithTimeout(ctx, 2*time.Second)

			f := funcs[i]
			done := make(chan error, 1)

			log.Info("closer: closing resource", zap.String("name", f.name))

			go func() {
				done <- f.fn()
			}()

			var err error

			select {
			case err = <-done:
			case <-rCtx.Done():
				err = rCtx.Err()
			}

			if err != nil {
				errs = append(errs,
					fmt.Errorf("closer: failed to close %q resource: %w", f.name, err),
				)
			} else {
				log.Info("closer: resource closed", zap.String("name", f.name))
			}

			rCancel()
		}

		result = errors.Join(errs...)
	})

	return result
}
