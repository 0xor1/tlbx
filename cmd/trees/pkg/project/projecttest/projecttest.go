package projecttest

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/consts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/test"
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

	p := (&project.Create{
		Base: project.Base{
			CurrencyCode: "USD",
			HoursPerDay:  8,
			DaysPerWeek:  5,
			StartOn:      ptr.Time(app.ExampleTime()),
			DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
			IsPublic:     false,
		},
		Name: "My New Project",
	}).MustDo(r.Ali().Client())

	p = (&project.One{Host: r.Ali().ID(), ID: p.ID}).MustDo(r.Ali().Client())

	p = (&project.Get{
		Host:           r.Ali().ID(),
		NameStartsWith: ptr.String("My New"),
		IsArchived:     false,
		IsPublic:       ptr.Bool(false),
		CreatedOnMin:   &p.CreatedOn,
		CreatedOnMax:   &p.CreatedOn,
		StartOnMin:     ptr.Time(app.ExampleTime()),
		StartOnMax:     ptr.Time(app.ExampleTime()),
		DueOnMin:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		DueOnMax:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		After:          nil,
		Sort:           consts.SortDueOn,
		Asc:            ptr.Bool(false),
		Limit:          ptr.Int(100),
	}).MustDo(r.Ali().Client()).Set[0]

	a.Zero(len((&project.Get{
		Host:           p.ID,
		NameStartsWith: ptr.String("My New"),
		IsArchived:     false,
		IsPublic:       ptr.Bool(false),
		CreatedOnMin:   &p.CreatedOn,
		CreatedOnMax:   &p.CreatedOn,
		StartOnMin:     ptr.Time(app.ExampleTime()),
		StartOnMax:     ptr.Time(app.ExampleTime()),
		DueOnMin:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		DueOnMax:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
		After:          ptr.ID(p.ID),
		Sort:           consts.SortDueOn,
		Asc:            ptr.Bool(true),
		Limit:          ptr.Int(100),
	}).MustDo(r.Ali().Client()).Set))
}
