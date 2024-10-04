package closer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Closer паттерн доводчик
type Closer struct {
	mux   sync.Mutex
	funcs []Func
}

type Func func(ctx context.Context) error

// Add регистрирует обработчик
func (c *Closer) Add(f Func) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.funcs = append(c.funcs, f)
}

// Close вызывает все операции по завершению программы
func (c *Closer) Close(ctx context.Context) error {
	// блокировка нужна для того, чтобы не было возможности зарегистрировать новый обработчик после вызова Close
	c.mux.Lock()
	defer c.mux.Unlock()

	msgs := make([]string, 0, len(c.funcs))
	complete := make(chan struct{}, 1)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	// в цикле конкурентно вызываются все функции завершения
	for _, fn := range c.funcs {
		wg.Add(1)
		go func(fn Func) {
			defer wg.Done()
			if err := fn(ctx); err != nil {
				mu.Lock()
				msgs = append(msgs, fmt.Sprintf("closer: %v", err))
				mu.Unlock()
			}
		}(fn)
	}
	go func() {
		wg.Wait()
		complete <- struct{}{}
	}()

	// в главной go-рутине мы ждем одного из двух событий: либо все завершат свою работу, либо выход из функции принудительно с ошибкой
	select {
	case <-complete:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %v", ctx.Err())
	}

	if len(msgs) > 0 {
		return fmt.Errorf("shutdown fubusged wutg error(s): \n%v", strings.Join(msgs, "\n"))
	}

	return nil
}
