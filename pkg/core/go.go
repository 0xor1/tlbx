package core

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
)

func Panic(v interface{}) {
	PanicOn(ToError(v))
}

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
			return fmt.Errorf("type: %T, value: %#v", v, v)
		}
	}
	return nil
}

func Recover(r func(err error)) {
	if err := ToError(recover()); err != nil {
		r(err)
	}
}

func Go(f func(), r func(err error)) {
	PanicIf(f == nil, "f must be none nil go routine func")
	PanicIf(r == nil, "r must be none nil recover func")
	go func() {
		defer Recover(r)
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
				gg.done(nil)
			}, gg.done)
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
	return fmt.Sprintf("error: %s\nstacktrace: %s\n", e.err.Error(), e.stackTrace)
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
	errs := make([]string, 0, len(gg.errs))
	for _, err := range gg.errs {
		errs = append(errs, err.Error())
	}
	return fmt.Sprintf("errors:\n%s", strings.Join(errs, ""))
}

func (gg *goGroup) String() string {
	return gg.Error()
}

func (gg *goGroup) done(e error) {
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
