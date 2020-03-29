package core

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_PanicOn(t *testing.T) {
	a := assert.New(t)
	PanicOn(nil)
	defer func() {
		a.Equal(assert.AnError.Error(), ToError(recover()).(*err).Message())
	}()

	PanicOn(assert.AnError)
}

func Test_PanicIf(t *testing.T) {
	a := assert.New(t)
	errStr := assert.AnError.Error()
	PanicIf(false, errStr)
	defer func() {
		a.Equal(errStr, ToError(recover()).(*err).Message())
	}()

	PanicIf(true, errStr)
}

func Test_ToError(t *testing.T) {
	a := assert.New(t)
	e := assert.AnError
	a.Equal(e.Error(), ToError(assert.AnError).(*err).Message())
	e = ToError(assert.AnError)
	a.Equal(e, ToError(e))
	a.Equal("test", ToError("test").(*err).Message())
	a.Equal("type: int, value: 42", ToError(42).(*err).Message())
}

func Test_Go(t *testing.T) {
	a := assert.New(t)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	Go(func() {
		panic(assert.AnError)
	}, func(i interface{}) {
		defer wg.Done()
		a.Equal(assert.AnError.Error(), ToError(i).(*err).Message())
	})
	wg.Wait()
}

func Test_GoGroup(t *testing.T) {
	a := assert.New(t)
	e := GoGroup(func() {
		panic(0)
	}, func() {
		panic(1)
	}, func() {
		panic(2)
	})

	a.Equal(3, len(e.Value().(Errors)))
	idxIsPresent := []bool{false, false, false}
	for _, e := range e.Value().(Errors) {
		runes := []rune(strings.Split(e.Error(), "\n")[0])
		idx, _ := strconv.Atoi(string(runes[len(runes)-1:]))
		idxIsPresent[idx] = true
	}
	a.True(idxIsPresent[0] && idxIsPresent[1] && idxIsPresent[2])

	e = GoGroup(func() {
		time.Sleep(time.Second)
	}, func() {
		time.Sleep(time.Second)
	}, func() {
		time.Sleep(time.Second)
	})
	a.Nil(e)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	e = GoGroup(func() {
		panic(0)
	}, func() {
		panic(1)
	}, func() {
		select {
		case <-time.After(2 * time.Second):
			panic(2)
		case <-ctx.Done():
			panic(3)
		}
	})

	a.Equal(3, len(e.Value().(Errors)))
	idxIsPresent = []bool{false, false, false, false}
	for _, e := range e.Value().(Errors) {
		runes := []rune(strings.Split(e.Error(), "\n")[0])
		idx, _ := strconv.Atoi(string(runes[len(runes)-1:]))
		idxIsPresent[idx] = true
	}
	a.True(idxIsPresent[0] && idxIsPresent[1] && !idxIsPresent[2] && idxIsPresent[3])
	GoGroup(nil)
	a.NotEmpty(e.Error())
	a.NotEmpty(e.Value().(Errors).String())
	a.NotEmpty(e.Value().(Errors)[0].(*err).String())
	a.NotEmpty(e.Value().(Errors)[0].(*err).Message())
	a.NotEmpty(e.Value().(Errors)[0].(*err).StackTrace())

	GoGroup()
}

func Test_Recover(t *testing.T) {
	Recover(nil)
}

func Test_Exit(t *testing.T) {
	Exit = func(i int) {}
	ExitIf(true, "test")
}
