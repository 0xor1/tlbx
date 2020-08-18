package projecttest

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/consts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.Get(),
		projecteps.Eps,
		true,
		nil,
		projecteps.OnDelete,
		true,
		projecteps.OnSetSocials,
		consts.FileBucket)
	defer r.CleanUp()

	ac := r.Ali().Client()
	p1 := (&project.Create{
		Base: project.Base{
			CurrencyCode: "USD",
			HoursPerDay:  8,
			DaysPerWeek:  5,
			StartOn:      ptr.Time(app.ExampleTime()),
			DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
			IsPublic:     false,
		},
		Name: "My New Project",
	}).MustDo(ac)

	// call it with a new client -> none logged in user (will only return public projects)
	nilP := (&project.One{Host: r.Ali().ID(), ID: p1.ID}).MustDo(r.NewClient())
	a.Nil(nilP)

	p1 = (&project.One{Host: r.Ali().ID(), ID: p1.ID}).MustDo(ac)
	a.NotNil(p1)

	p1 = (&project.Get{
		Host:           r.Ali().ID(),
		NameStartsWith: ptr.String("My New"),
		IsArchived:     false,
		IsPublic:       ptr.Bool(false),
		CreatedOnMin:   &p1.CreatedOn,
		CreatedOnMax:   &p1.CreatedOn,
		StartOnMin:     ptr.Time(app.ExampleTime()),
		StartOnMax:     ptr.Time(app.ExampleTime()),
		DueOnMin:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		DueOnMax:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		After:          nil,
		Sort:           consts.SortDueOn,
		Asc:            ptr.Bool(false),
		Limit:          ptr.Int(100),
	}).MustDo(ac).Set[0]

	a.Zero(len((&project.Get{
		Host:           p1.ID,
		NameStartsWith: ptr.String("My New"),
		IsArchived:     false,
		IsPublic:       ptr.Bool(false),
		CreatedOnMin:   &p1.CreatedOn,
		CreatedOnMax:   &p1.CreatedOn,
		StartOnMin:     ptr.Time(app.ExampleTime()),
		StartOnMax:     ptr.Time(app.ExampleTime()),
		DueOnMin:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		DueOnMax:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		After:          ptr.ID(p1.ID),
		Sort:           consts.SortDueOn,
		Asc:            ptr.Bool(true),
		Limit:          ptr.Int(100),
	}).MustDo(ac).Set))

	name1 := "renamed project"
	cc := "EUR"
	dpw := uint8(4)
	startOn := NowMilli()
	dueOn := startOn.Add(24 * time.Hour)
	p1 = (&project.Update{
		ID:           p1.ID,
		Name:         &field.String{V: name1},
		CurrencyCode: &field.String{V: cc},
		HoursPerDay:  &field.UInt8{V: dpw},
		DaysPerWeek:  &field.UInt8{V: dpw},
		StartOn:      &field.TimePtr{V: &startOn},
		DueOn:        &field.TimePtr{V: &dueOn},
		IsArchived:   &field.Bool{V: false},
		IsPublic:     &field.Bool{V: true},
	}).MustDo(ac)
	a.Equal(name1, p1.Name)
	a.Equal(cc, p1.CurrencyCode)
	a.Equal(dpw, p1.HoursPerDay)
	a.Equal(dpw, p1.DaysPerWeek)
	a.Equal(startOn, *p1.StartOn)
	a.Equal(dueOn, *p1.DueOn)
	a.False(p1.IsArchived)
	a.True(p1.IsPublic)

	// call it with a new client -> none logged in user (will only return public projects)
	p1 = (&project.One{Host: r.Ali().ID(), ID: p1.ID}).MustDo(r.NewClient())
	a.NotNil(p1)

	// try to set startOn to same value as dueOn
	nilP, err := (&project.Update{
		ID:      p1.ID,
		StartOn: &field.TimePtr{V: &dueOn},
	}).Do(ac)
	a.Nil(nilP)
	a.Contains(err.Error(), "invalid startOn must be before dueOn")

	// try to set dueOn to same value as startOn
	nilP, err = (&project.Update{
		ID:    p1.ID,
		DueOn: &field.TimePtr{V: &startOn},
	}).Do(ac)
	a.Nil(nilP)
	a.Contains(err.Error(), "invalid startOn must be before dueOn")

	// create another project and get with a limit of 1 to test more: true response
	name2 := "My 2nd Project"
	p2 := (&project.Create{
		Base: project.Base{
			CurrencyCode: "USD",
			HoursPerDay:  8,
			DaysPerWeek:  5,
			IsPublic:     false,
		},
		Name: name2,
	}).MustDo(ac)
	a.NotNil(p2)

	a.True((&project.Get{
		Host:  r.Ali().ID(),
		Limit: ptr.Int(1),
	}).MustDo(ac).More)

	// trigger OnSetSocials code
	(&user.SetHandle{
		Handle: "ali_changed",
	}).MustDo(ac)

	// make empty update
	ps := (&project.Updates{}).MustDo(ac)

	// make multiple updates
	dueOn = dueOn.Add(24 * time.Hour)
	ps = (&project.Updates{
		{
			ID:    p1.ID,
			DueOn: &field.TimePtr{V: &dueOn},
		},
		{
			ID:    p2.ID,
			DueOn: &field.TimePtr{V: &dueOn},
		},
	}).MustDo(ac)
	a.Equal(name1, ps[0].Name)
	a.Equal(dueOn, *ps[0].DueOn)
	a.Equal(name2, ps[1].Name)
	a.Equal(dueOn, *ps[1].DueOn)

	// delete projects
	(&project.Delete{}).MustDo(ac)
	(&project.Delete{p1.ID, p2.ID}).MustDo(ac)
	a.Zero(len((&project.Get{Host: r.Ali().ID()}).MustDo(ac).Set))
}
