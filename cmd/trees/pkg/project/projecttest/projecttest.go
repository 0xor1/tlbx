package projecttest

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
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
		cnsts.TempFileBucket,
		cnsts.FileBucket)
	defer r.CleanUp()

	ac := r.Ali().Client()
	p1 := (&project.Create{
		CurrencyCode: "USD",
		HoursPerDay:  8,
		DaysPerWeek:  5,
		StartOn:      ptr.Time(app.ExampleTime()),
		DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		IsPublic:     false,
		Name:         "My New Project",
	}).MustDo(ac)

	// call it with a new client -> none logged in user (will only return public projects)
	nilP := (&project.One{Host: r.Ali().ID(), ID: p1.ID}).MustDo(r.NewClient())
	a.Nil(nilP)

	p1 = (&project.One{Host: r.Ali().ID(), ID: p1.ID}).MustDo(ac)
	a.NotNil(p1)

	p1 = (&project.Get{
		Host:         ptr.ID(r.Ali().ID()),
		NamePrefix:   ptr.String("My New"),
		IsArchived:   false,
		IsPublic:     ptr.Bool(false),
		CreatedOnMin: &p1.CreatedOn,
		CreatedOnMax: &p1.CreatedOn,
		StartOnMin:   ptr.Time(app.ExampleTime()),
		StartOnMax:   ptr.Time(app.ExampleTime()),
		DueOnMin:     ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		DueOnMax:     ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		After:        nil,
		Sort:         cnsts.SortDueOn,
		Asc:          ptr.Bool(false),
		Limit:        100,
	}).MustDo(ac).Set[0]

	// getwithout specifying host
	p1 = (&project.Get{
		NamePrefix:   ptr.String("My New"),
		IsArchived:   false,
		IsPublic:     ptr.Bool(false),
		CreatedOnMin: &p1.CreatedOn,
		CreatedOnMax: &p1.CreatedOn,
		StartOnMin:   ptr.Time(app.ExampleTime()),
		StartOnMax:   ptr.Time(app.ExampleTime()),
		DueOnMin:     ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		DueOnMax:     ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		After:        nil,
		Sort:         cnsts.SortDueOn,
		Asc:          ptr.Bool(false),
		Limit:        100,
	}).MustDo(ac).Set[0]

	a.Zero(len((&project.Get{
		Host:         ptr.ID(p1.ID),
		NamePrefix:   ptr.String("My New"),
		IsArchived:   false,
		IsPublic:     ptr.Bool(false),
		CreatedOnMin: &p1.CreatedOn,
		CreatedOnMax: &p1.CreatedOn,
		StartOnMin:   ptr.Time(app.ExampleTime()),
		StartOnMax:   ptr.Time(app.ExampleTime()),
		DueOnMin:     ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		DueOnMax:     ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		After:        ptr.ID(p1.ID),
		Sort:         cnsts.SortDueOn,
		Asc:          ptr.Bool(true),
		Limit:        100,
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
		CurrencyCode: "USD",
		HoursPerDay:  8,
		DaysPerWeek:  5,
		IsPublic:     false,
		Name:         name2,
	}).MustDo(ac)
	a.NotNil(p2)

	a.True((&project.Get{
		Host:  ptr.ID(r.Ali().ID()),
		Limit: 1,
	}).MustDo(ac).More)

	// trigger OnSetSocials code
	(&user.SetHandle{
		Handle: "ali_" + r.Unique(),
	}).MustDo(ac)

	// make empty update
	ps := (&project.Updates{}).MustDo(ac)

	// make multiple updates
	dueOn = dueOn.Add(24 * time.Hour)
	ps = (&project.Updates{
		{
			// send an empty update
			ID: p1.ID,
		},
		{
			ID:    p1.ID,
			DueOn: &field.TimePtr{V: &dueOn},
		},
		{
			ID:    p2.ID,
			DueOn: &field.TimePtr{V: &dueOn},
		},
	}).MustDo(ac)
	a.Len(ps, 2)
	a.True(p1.ID.Equal(ps[0].ID))
	a.Equal(name1, ps[0].Name)
	a.Equal(dueOn, *ps[0].DueOn)
	a.True(p2.ID.Equal(ps[1].ID))
	a.Equal(name2, ps[1].Name)
	a.Equal(dueOn, *ps[1].DueOn)

	// delete projects
	(&project.Delete{}).MustDo(ac)
	(&project.Delete{p1.ID, p2.ID}).MustDo(ac)
	a.Zero(len((&project.Get{Host: ptr.ID(r.Ali().ID())}).MustDo(ac).Set))

	p1 = (&project.Create{
		CurrencyCode: "USD",
		HoursPerDay:  8,
		DaysPerWeek:  5,
		StartOn:      ptr.Time(app.ExampleTime()),
		DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		IsPublic:     false,
		Name:         "My New Project",
	}).MustDo(ac)

	// test empty request
	(&project.AddUsers{
		Host:    r.Ali().ID(),
		Project: p1.ID,
	}).MustDo(ac)

	(&project.AddUsers{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		Users: []*project.SendUser{
			{
				ID:   r.Bob().ID(),
				Role: cnsts.RoleAdmin,
			},
			{
				ID:   r.Cat().ID(),
				Role: cnsts.RoleWriter,
			},
			{
				ID:   r.Dan().ID(),
				Role: cnsts.RoleReader,
			},
		},
	}).MustDo(ac)

	us := (&project.GetUsers{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		IDs: IDs{
			r.Cat().ID(),
			r.Dan().ID(),
		},
	}).MustDo(r.Dan().Client())
	a.False(us.More)
	a.Len(us.Set, 2)
	a.True(us.Set[0].ID.Equal(r.Cat().ID()))

	role := cnsts.RoleWriter
	us = (&project.GetUsers{
		Host:         r.Ali().ID(),
		Project:      p1.ID,
		Role:         &role,
		HandlePrefix: ptr.String("ca"),
	}).MustDo(r.Dan().Client())
	a.False(us.More)
	a.Len(us.Set, 1)
	a.True(us.Set[0].ID.Equal(r.Cat().ID()))

	us = (&project.GetUsers{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		After:   ptr.ID(r.Ali().ID()),
		Limit:   2,
	}).MustDo(r.Dan().Client())
	a.True(us.More)
	a.Len(us.Set, 2)
	a.True(us.Set[0].ID.Equal(r.Bob().ID()))
	a.Equal(us.Set[0].Role, cnsts.RoleAdmin)
	a.True(us.Set[1].ID.Equal(r.Cat().ID()))
	a.Equal(us.Set[1].Role, cnsts.RoleWriter)

	// send empty req
	(&project.SetUserRoles{
		Host:    r.Ali().ID(),
		Project: p1.ID,
	}).MustDo(ac)

	(&project.SetUserRoles{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		Users: []*project.SendUser{
			{
				ID:   r.Bob().ID(),
				Role: cnsts.RoleReader,
			},
			{
				ID:   r.Cat().ID(),
				Role: cnsts.RoleReader,
			},
			{
				ID:   r.Dan().ID(),
				Role: cnsts.RoleAdmin,
			},
		},
	}).MustDo(ac)

	us = (&project.GetUsers{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		After:   ptr.ID(r.Ali().ID()),
	}).MustDo(r.Dan().Client())
	a.False(us.More)
	a.Len(us.Set, 3)
	a.True(us.Set[0].ID.Equal(r.Dan().ID()))
	a.Equal(us.Set[0].Role, cnsts.RoleAdmin)
	a.True(us.Set[1].ID.Equal(r.Bob().ID()))
	a.Equal(us.Set[1].Role, cnsts.RoleReader)
	a.True(us.Set[2].ID.Equal(r.Cat().ID()))
	a.Equal(us.Set[2].Role, cnsts.RoleReader)

	me := (&project.GetMe{
		Host:    r.Ali().ID(),
		Project: p1.ID,
	}).MustDo(r.Dan().Client())
	a.True(me.ID.Equal(r.Dan().ID()))

	// send empty req
	(&project.RemoveUsers{
		Host:    r.Ali().ID(),
		Project: p1.ID,
	}).MustDo(r.Dan().Client())
	a.True(me.ID.Equal(r.Dan().ID()))

	(&project.RemoveUsers{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		Users:   IDs{r.Bob().ID()},
	}).MustDo(r.Dan().Client())
	a.True(me.ID.Equal(r.Dan().ID()))

	me, err = (&project.GetMe{
		Host:    r.Ali().ID(),
		Project: p1.ID,
	}).Do(r.Bob().Client())
	a.Nil(me)
	a.Contains(err.Error(), "Forbidden")

	// dan removed one user => only 1 activity
	as := (&project.GetActivities{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		User:    ptr.ID(r.Dan().ID()),
	}).MustDo(ac)
	a.Len(as.Set, 1)
	a.False(as.More)

	as = (&project.GetActivities{
		Host:    r.Ali().ID(),
		Project: p1.ID,
		Item:    ptr.ID(p1.ID),
	}).MustDo(ac)
	a.Len(as.Set, 1)
	a.False(as.More)

	as = (&project.GetActivities{
		Host:    r.Ali().ID(),
		Project: p1.ID,
	}).MustDo(ac)
	a.Len(as.Set, 8)
	a.False(as.More)
	item1OccuredOn := as.Set[0].OccurredOn
	item2OccuredOn := as.Set[1].OccurredOn

	as = (&project.GetActivities{
		Host:          r.Ali().ID(),
		Project:       p1.ID,
		OccuredBefore: &as.Set[0].OccurredOn,
		Limit:         2,
	}).MustDo(ac)
	a.Equal(item2OccuredOn, as.Set[0].OccurredOn)
	a.Len(as.Set, 2)
	a.True(as.More)

	as = (&project.GetActivities{
		Host:         r.Ali().ID(),
		Project:      p1.ID,
		OccuredAfter: &as.Set[0].OccurredOn,
		Limit:        2,
	}).MustDo(ac)
	a.Equal(item1OccuredOn, as.Set[0].OccurredOn)
	a.Len(as.Set, 1)
	a.False(as.More)
}
