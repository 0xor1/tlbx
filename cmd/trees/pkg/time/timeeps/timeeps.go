package timeeps

import (
	"net/http"
	time_ "time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/epsutil"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/time"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new time",
			Path:         (&time.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &time.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &time.Create{
					Host:     app.ExampleID(),
					Project:  app.ExampleID(),
					Task:     app.ExampleID(),
					Duration: 60,
					Note:     "I did an hours work",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleTime
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*time.Create)
				me := me.Get(tlbx)
				app.ReturnIf(args.Duration == 0 || args.Duration > durationMax, http.StatusBadRequest, "duration must be between 1 and 1440")
				args.Note = StrTrimWS(args.Note)
				validate.Str("note", args.Note, tlbx, 0, noteMaxLen)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				epsutil.MustLockProject(tx, args.Host, args.Project)
				task := taskeps.One(tx, args.Host, args.Project, args.Task)
				app.ReturnIf(task == nil, http.StatusNotFound, "task not found")
				t := &time.Time{
					Task:      args.Task,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: NowMilli(),
					Duration:  args.Duration,
					Note:      args.Note,
				}
				// insert new time
				_, err := tx.Exec(`INSERT INTO times (host, project, task, id, createdBy, createdOn, duration, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, args.Task, t.ID, t.CreatedBy, t.CreatedOn, t.Duration, t.Note)
				PanicOn(err)
				// update task loggedTime
				_, err = tx.Exec(`UPDATE tasks SET loggedTime=loggedTime+? WHERE host=? AND project=? AND id=?`, args.Duration, args.Host, args.Project, args.Task)
				PanicOn(err)
				// propogate aggregate values upwards
				epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, args.Task)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, t.ID, cnsts.TypeTime, cnsts.ActionCreated, &task.Name, nil, args)
				tx.Commit()
				return t
			},
		},
		{
			Description:  "Update a time",
			Path:         (&time.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &time.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &time.Update{
					Host:     app.ExampleID(),
					Project:  app.ExampleID(),
					Task:     app.ExampleID(),
					ID:       app.ExampleID(),
					Duration: &field.UInt64{V: 60},
					Note:     &field.String{V: "woo"},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleTime
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*time.Update)
				me := me.Get(tlbx)
				if args.Duration == nil &&
					args.Note == nil {
					// nothing to update
					return nil
				}
				app.ReturnIf(args.Duration != nil && (args.Duration.V == 0 || args.Duration.V > durationMax), http.StatusBadRequest, "duration must be between 1 and 1440")
				if args.Note != nil {
					args.Note.V = StrTrimWS(args.Note.V)
					validate.Str("note", args.Note.V, tlbx, 0, noteMaxLen)
				}
				app.ReturnIf(args.Duration != nil && (args.Duration.V == 0 || args.Duration.V > durationMax), http.StatusBadRequest, "duration must be between 1 and 1440")
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				t := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "time entry not found")
				app.ReturnIf(!t.CreatedBy.Equal(me) && role != cnsts.RoleAdmin, http.StatusForbidden, "")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.CreatedOn.Before(Now().Add(-1*time_.Hour))), http.StatusForbidden, "you may only edit your own time entries within an hour of creating it")
				treeUpdate := false
				var diff uint64
				var sign string
				if args.Duration != nil && args.Duration.V != t.Duration {
					if args.Duration.V > t.Duration {
						diff = args.Duration.V - t.Duration
						sign = "+"
					} else {
						diff = t.Duration - args.Duration.V
						sign = "-"
					}
					t.Duration = args.Duration.V
					treeUpdate = true
				}
				if args.Note != nil && args.Note.V != t.Note {
					t.Note = args.Note.V
				}
				_, err := tx.Exec(`UPDATE times SET duration=?, note=? WHERE host=? AND project=? AND task=? AND id=?`, t.Duration, t.Note, args.Host, args.Project, t.Task, t.ID)
				PanicOn(err)
				if treeUpdate {
					_, err = tx.Exec(Strf(`UPDATE tasks SET loggedTime=loggedTime%s? WHERE host=? AND project=? AND id=?`, sign), diff, args.Host, args.Project, t.Task)
					PanicOn(err)
					epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, args.Task)
				}
				tx.Commit()
				return t
			},
		},
		{
			Description:  "Delete time",
			Path:         (&time.Delete{}).Path(),
			Timeout:      5000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &time.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &time.Delete{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					ID:      app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				//args := a.(*time.Delete)

				return nil
			},
		},
	}

	durationMax uint64 = 1440 // 24hrs in mins
	noteMaxLen         = 250
	exampleTime        = &time.Time{
		Task:      app.ExampleID(),
		ID:        app.ExampleID(),
		CreatedBy: app.ExampleID(),
		CreatedOn: app.ExampleTime(),
		Duration:  60,
		Note:      "I did something",
	}
)

func getOne(tx service.Tx, host, project, task, id ID) *time.Time {
	row := tx.QueryRow(`SELECT task, id, createdBy, createdOn, duration, note FROM times WHERE host=? AND project=? AND task=? AND id=?`, host, project, task, id)
	t, err := Scan(row)
	sql.PanicIfIsntNoRows(err)
	return t
}

func Scan(r isql.Row) (*time.Time, error) {
	t := &time.Time{}
	err := r.Scan(
		&t.Task,
		&t.ID,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.Duration,
		&t.Note)
	if t.ID.IsZero() {
		t = nil
	}
	return t, err
}
