package taskeps

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/epsutil"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new task",
			Path:         (&task.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Create{
					Host:            app.ExampleID(),
					Project:         app.ExampleID(),
					Parent:          app.ExampleID(),
					PreviousSibling: ptr.ID(app.ExampleID()),
					Name:            "do it",
					Description:     ptr.String("do the thing you're supposed to do"),
					IsParallel:      true,
					User:            ptr.ID(app.ExampleID()),
					EstimatedTime:   40,
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Create)
				me := me.Get(tlbx)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				if args.Description != nil && *args.Description == "" {
					args.Description = nil
				}
				if args.Description != nil {
					validate.Str("description", *args.Description, tlbx, descriptionMinLen, descriptionMaxLen)
				}
				epsutil.MustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				t := &task.Task{
					ID:                  tlbx.NewID(),
					Parent:              &args.Parent,
					FirstChild:          nil,
					NextSibling:         nil,
					User:                args.User,
					Name:                args.Name,
					Description:         args.Description,
					CreatedBy:           me,
					CreatedOn:           NowMilli(),
					MinimumTime:         0,
					EstimatedTime:       0,
					LoggedTime:          0,
					EstimatedSubTime:    0,
					LoggedSubTime:       0,
					EstimatedExpense:    0,
					LoggedExpense:       0,
					EstimatedSubExpense: 0,
					LoggedSubExpense:    0,
					FileCount:           0,
					FileSize:            0,
					SubFileCount:        0,
					SubFileSize:         0,
					ChildCount:          0,
					DescendantCount:     0,
					IsParallel:          args.IsParallel,
				}
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				// lock project, required for any action that will change aggregate values
				epsutil.MustLockProject(tlbx, tx, args.Host, args.Project)
				// get full ancestor list, for updating aggregate values,
				//tx.Query()
				// TODO get previous and next Siblings for updating pointers,
				// TODO validate args.User is an active user with write or admin permissions
				_, err := tx.Exec(`INSERT INTO tasks (host, project, id, parent, firstChild, nextSibling, user, name, description, isParallel, createdBy, createdOn, minimumTime, estimatedTime, loggedTime, estimatedSubTime, loggedSubTime, estimatedExpense, loggedExpense, estimatedSubExpense, loggedSubExpense, fileCount, fileSize, subFileCount, subFileSize, childCount, descendantCount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, t.ID, t.Parent, t.FirstChild, t.NextSibling, t.User, t.Name, t.Description, t.IsParallel, t.CreatedBy, t.CreatedOn, t.MinimumTime, t.EstimatedTime, t.LoggedTime, t.EstimatedSubTime, t.LoggedSubTime, t.EstimatedExpense, t.LoggedExpense, t.EstimatedSubExpense, t.LoggedSubExpense, t.FileCount, t.FileSize, t.SubFileCount, t.SubFileSize, t.ChildCount, t.DescendantCount)
				PanicOn(err)
				return nil
			},
		},
	}

	nameMinLen        = 1
	nameMaxLen        = 250
	descriptionMinLen = 1
	descriptionMaxLen = 1250
	exampleTask       = &task.Task{
		ID:                  app.ExampleID(),
		Parent:              ptr.ID(app.ExampleID()),
		FirstChild:          ptr.ID(app.ExampleID()),
		NextSibling:         ptr.ID(app.ExampleID()),
		User:                ptr.ID(app.ExampleID()),
		Name:                "do it",
		Description:         ptr.String("do that thing you're supposed to do"),
		CreatedBy:           app.ExampleID(),
		CreatedOn:           app.ExampleTime(),
		MinimumTime:         100,
		EstimatedTime:       100,
		LoggedTime:          100,
		EstimatedSubTime:    100,
		LoggedSubTime:       100,
		EstimatedExpense:    100,
		LoggedExpense:       100,
		EstimatedSubExpense: 100,
		LoggedSubExpense:    100,
		FileCount:           100,
		FileSize:            100,
		SubFileCount:        100,
		SubFileSize:         100,
		ChildCount:          100,
		DescendantCount:     100,
		IsParallel:          true,
	}
)
