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
		a.Equal(assert.AnError, ToError(recover()))
	}()

	PanicOn(assert.AnError)
}

func Test_PanicIf(t *testing.T) {
	a := assert.New(t)
	errStr := assert.AnError.Error()
	PanicIf(false, errStr)
	defer func() {
		a.Equal(errStr, ToError(recover()).Error())
	}()

	PanicIf(true, errStr)
}

func Test_ToError(t *testing.T) {
	a := assert.New(t)
	defer func() {
		a.Equal(assert.AnError, ToError(recover()))
		defer func() {
			a.Equal("recover type: int, value: 1", ToError(recover()).Error())
		}()
		panic(1)
	}()
	panic(assert.AnError)
}

func Test_Go(t *testing.T) {
	a := assert.New(t)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	Go(func() {
		panic(assert.AnError)
	}, func(err error) {
		defer wg.Done()
		a.Equal(assert.AnError, err)
	})
	wg.Wait()
}

func Test_GoGroup(t *testing.T) {
	a := assert.New(t)
	err := GoGroup(func() {
		panic(0)
	}, func() {
		panic(1)
	}, func() {
		panic(2)
	})

	a.Equal(3, len(err.(*goGroup).errs))
	idxIsPresent := []bool{false, false, false}
	for _, err := range err.(*goGroup).errs {
		runes := []rune(strings.Split(err.Error(), "\n")[1])
		idx, _ := strconv.Atoi(string(runes[len(runes)-1:]))
		idxIsPresent[idx] = true
	}
	a.True(idxIsPresent[0] && idxIsPresent[1] && idxIsPresent[2])

	a.Nil(GoGroup(func() {
		time.Sleep(time.Second)
	}, func() {
		time.Sleep(time.Second)
	}, func() {
		time.Sleep(time.Second)
	}))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	err = GoGroup(func() {
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

	a.Equal(3, len(err.(*goGroup).errs))
	idxIsPresent = []bool{false, false, false, false}
	for _, e := range err.(*goGroup).errs {
		runes := []rune(strings.Split(e.Error(), "\n")[1])
		idx, _ := strconv.Atoi(string(runes[len(runes)-1:]))
		idxIsPresent[idx] = true
	}
	a.True(idxIsPresent[0] && idxIsPresent[1] && !idxIsPresent[2] && idxIsPresent[3])
	a.Nil(GoGroup())
	MustGoGroup()
	a.NotEmpty(err.Error())
	a.NotEmpty(err.(*goGroup).String())
	a.NotEmpty(err.(*goGroup).errs[0].String())
}
