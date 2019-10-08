package core

import (
	"fmt"
	"runtime/debug"
	"sync"
)

func PanicOn(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIf(condition bool, format string, args ...interface{}) {
	if condition {
		PanicOn(fmt.Errorf(format, args...))
	}
}

func ToError(v interface{}) error {
	if v != nil {
		if err, ok := v.(error); ok {
			return err
		} else {
			return fmt.Errorf("recover type: %T, value: %#v", v, v)
		}
	}
	return nil
}

func Go(f func(), r func(err error)) {
	PanicIf(f == nil, "f must be none nil go routine func")
	PanicIf(r == nil, "r must be none nil recover func")
	go func() {
		defer func() {
			if err := ToError(recover()); err != nil {
				r(err)
			}
		}()
		f()
	}()
}

func GoGroup(fs ...func()) error {
	if len(fs) == 0 {
		return nil
	}
	gg := &goGroup{
		errs:    make([]*stackError, 0, len(fs)),
		errsMtx: &sync.Mutex{},
		wg:      &sync.WaitGroup{},
	}
	gg.wg.Add(len(fs))
	for _, f := range fs {
		func(f func()) {
			Go(func() {
				f()
				gg.Done(nil)
			}, gg.Done)
		}(f)
	}
	gg.wg.Wait()
	if len(gg.errs) > 0 {
		return gg
	}
	return nil
}

func MustGoGroup(fs ...func()) {
	PanicOn(GoGroup(fs...))
}

type stackError struct {
	err        error
	stackTrace string
}

func (e *stackError) Error() string {
	return fmt.Sprintf("error:\n%s\nstacktrace:\n%s", e.err.Error(), e.stackTrace)
}

func (e *stackError) String() string {
	return e.Error()
}

type goGroup struct {
	errs    []*stackError
	errsMtx *sync.Mutex
	wg      *sync.WaitGroup
}

func (gg *goGroup) Error() string {
	return fmt.Sprintf("errors:\n%s\n", gg.errs)
}

func (gg *goGroup) String() string {
	return gg.Error()
}

func (gg *goGroup) Done(e error) {
	defer gg.wg.Done()
	if e != nil {
		gg.errsMtx.Lock()
		defer gg.errsMtx.Unlock()
		gg.errs = append(gg.errs, &stackError{
			err:        e,
			stackTrace: string(debug.Stack()),
		})
	}
}
