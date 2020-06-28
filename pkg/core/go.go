package core

import (
	"os"
	"runtime/debug"
	"strings"
	"sync"
)

var (
	Exit = os.Exit
)

type Error interface {
	Error() string
	String() string
	Message() string
	StackTrace() string
	Value() interface{}
}

type Errors []Error

func (es Errors) Error() string {
	errStrs := make([]string, 0, len(es))
	for _, err := range es {
		errStrs = append(errStrs, err.Error())
	}
	return Sprintf("errors:\n%s", strings.Join(errStrs, ""))
}

func (es Errors) String() string {
	return es.Error()
}

type err struct {
	Message_    string      `json:"message"`
	StackTrace_ string      `json:"stackTrace"`
	Value_      interface{} `json:"value"`
}

func (e *err) Error() string {
	return Sprintf("message: %s\nstackTrace: %s", e.Message_, e.StackTrace_)
}

func (e *err) String() string {
	return e.Error()
}

func (e *err) Message() string {
	return e.Message_
}

func (e *err) StackTrace() string {
	return e.StackTrace_
}

func (e *err) Value() interface{} {
	return e.Value_
}

func ToError(i interface{}) Error {
	if i == nil {
		return nil
	}

	var msg string

	switch v := i.(type) {
	case *err:
		return v
	case error:
		msg = v.Error()
	case string:
		msg = v
	default:
		msg = Sprintf("type: %T, value: %#v", i, i)
	}

	return &err{
		Message_:    msg,
		StackTrace_: string(debug.Stack()),
		Value_:      i,
	}
}

func ExitOn(i interface{}) {
	if i != nil {
		Exit(1)
	}
}

func ExitIf(condition bool, format string, args ...interface{}) {
	if condition {
		ExitOn(Sprintf(format, args...))
	}
}

func PanicOn(i interface{}) {
	if err := ToError(i); err != nil {
		panic(err)
	}
}

func PanicIf(condition bool, format string, args ...interface{}) {
	if condition {
		PanicOn(Sprintf(format, args...))
	}
}

func Recover(r func(interface{})) {
	if r == nil {
		return
	}
	if err := ToError(recover()); err != nil {
		r(err)
	}
}

func Do(f func(), r func(interface{})) {
	defer Recover(r)
	f()
}

func Go(f func(), r func(interface{})) {
	go Do(f, r)
}

func GoGroup(fs ...func()) Error {
	if len(fs) == 0 {
		return nil
	}
	gg := &goGroup{
		errs:    make(Errors, 0, len(fs)),
		errsMtx: &sync.Mutex{},
		wg:      &sync.WaitGroup{},
	}
	gg.wg.Add(len(fs))
	for _, a := range fs {
		func(f func()) {
			Go(func() {
				f()
				gg.done(nil)
			}, gg.done)
		}(a)
	}
	gg.wg.Wait()
	if len(gg.errs) > 0 {
		return ToError(gg.errs)
	}
	return nil
}

func MustGoGroup(fs ...func()) {
	PanicOn(GoGroup(fs...))
}

type goGroup struct {
	errs    Errors
	errsMtx *sync.Mutex
	wg      *sync.WaitGroup
}

func (gg *goGroup) done(i interface{}) {
	defer gg.wg.Done()
	if i != nil {
		gg.errsMtx.Lock()
		defer gg.errsMtx.Unlock()
		gg.errs = append(gg.errs, ToError(i))
	}
}
