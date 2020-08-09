package projecteps

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new project",
			Path:         (&project.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Create{
					Base: project.Base{
						CurrencyCode: "USD",
						HoursPerDay:  8,
						DaysPerWeek:  5,
						IsPublic:     false,
					},
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Create{
					Base: project.Base{
						CurrencyCode: "USD",
						HoursPerDay:  8,
						DaysPerWeek:  5,
						StartOn:      ptr.Time(app.ExampleTime()),
						DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
						IsPublic:     false,
					},
					Name: "My New Project",
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.Project{
					Task: task.Task{
						Name: "My New Project",
					},
					Base: project.Base{
						HoursPerDay: 8,
						DaysPerWeek: 5,
						StartOn:     ptr.Time(app.ExampleTime()),
						DueOn:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
						IsPublic:    false,
					},
					IsArchived: false,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.Create)
				me := me.Get(tlbx)
				validate.Str("name", args.Name, tlbx, 1, nameMaxLen)
				tlbx.BadReqIf(args.HoursPerDay < 1 || args.HoursPerDay > 24, "invalid hoursPerDay must be > 0 and <= 24")
				tlbx.BadReqIf(args.DaysPerWeek < 1 || args.HoursPerDay > 7, "invalid daysPerWeek must be > 0 and <= 7")
				tlbx.BadReqIf(args.StartOn != nil && args.DueOn != nil && args.StartOn.After(*args.Base.DueOn), "invalid startOn must be before dueOn")
				project := &project.Project{
					Task: task.Task{
						ID:         tlbx.NewID(),
						Name:       args.Name,
						CreatedOn:  NowMilli(),
						IsParallel: true,
					},
					Base:       args.Base,
					IsArchived: false,
				}
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				tx.Exec(`INSERT INTO projectLocks (host, id) VALUES (?, ?)`, me, project.ID)
				tx.Exec(`INSERT INTO projects (host, id, isArchived, name, createdOn, hoursPerDay, daysPerWeek, startOn, dueOn, isPublic) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, me, project.ID, project.IsArchived, project.Name, project.CreatedOn, project.HoursPerDay, project.DaysPerWeek, project.StartOn, project.DueOn, project.IsPublic)
				// TODO continue from here tomorrow
				tx.Exec(`INSERT INTO tasks (host, project, id, parent, firstChild, nextSibling, user, name, description, createdOn, minimumRemainingTime, estimatedTime, loggedTime, estimatedSubTime, loggedSubTime, estimatedCost, loggedCost, estimatedSubCost, loggedSubCost, fileCount, fileSize, subFileCount, subFileSize, childCount, descendantCount, isParallel) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, me, project.ID, project.ID, project.Name, project.CreatedOn, project.HoursPerDay, project.DaysPerWeek, project.StartOn, project.DueOn, project.IsPublic)
				return nil
			},
		},
	}
	nameMaxLen  = 250
	aliasMaxLen = 50
)

func OnActivate(tlbx app.Tlbx, me *user.User) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	// _, err := tx.Exec(`INSERT INTO accounts WHERE id=?`, me)
	// PanicOn(err)
	tx.Commit()
}

func OnDelete(tlbx app.Tlbx, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	// TODO delete all files from minio/s3
	defer tx.Rollback()
	_, err := tx.Exec(`DELETE FROM projectLocks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projectUsers WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projectActivities WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projects WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM tasks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM timeLogs WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM files WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM comments WHERE host=?`, me)
	PanicOn(err)
	tx.Commit()
}
