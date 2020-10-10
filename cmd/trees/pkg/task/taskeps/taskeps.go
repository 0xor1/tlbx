package taskeps

import (
	"net/http"
	"strings"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/epsutil"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
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
				return exampleTask
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
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
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
					MinimumTime:         args.EstimatedTime,
					EstimatedTime:       args.EstimatedTime,
					LoggedTime:          0,
					EstimatedSubTime:    0,
					LoggedSubTime:       0,
					EstimatedExpense:    args.EstimatedExpense,
					LoggedExpense:       0,
					EstimatedSubExpense: 0,
					LoggedSubExpense:    0,
					FileCount:           0,
					FileSize:            0,
					FileSubCount:        0,
					FileSubSize:         0,
					ChildCount:          0,
					DescendantCount:     0,
					IsParallel:          args.IsParallel,
				}
				if args.User != nil && !args.User.Equal(me) {
					// if Im assigning to someone that isnt me,
					// validate that user has write access to this
					// project
					epsutil.MustHaveAccess(tlbx, args.Host, args.Project, args.User, cnsts.RoleWriter)
				}
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				// lock project, required for any action that will change aggregate values nad/or tree structure
				epsutil.MustLockProject(tlbx, tx, args.Host, args.Project)
				// get correct next sibling value from either previousSibling if
				// specified or parent.FirstChild otherwise. Then update previousSiblings nextSibling value
				// or parents firstChild value depending on the scenario.
				var previousSibling *task.Task
				if args.PreviousSibling != nil {
					previousSibling = getOne(tlbx, tx, args.Host, args.Project, *args.PreviousSibling)
					t.NextSibling = previousSibling.NextSibling
					previousSibling.NextSibling = &t.ID
					// point previous sibling at new task
					_, err := tx.Exec(`UPDATE tasks SET nextSibling=? WHERE host=? AND project=? AND id=?`, t.ID, args.Host, args.Project, previousSibling.ID)
					PanicOn(err)
				} else {
					// else newTask is being inserted as firstChild, so set any current firstChild
					// as newTask's NextSibling
					// get parent for updating child/descendant counts and firstChild if required
					parent := getOne(tlbx, tx, args.Host, args.Project, args.Parent)
					t.NextSibling = parent.FirstChild
					// increment parents child and descendant counters and firstChild pointer incase that was changed
					_, err := tx.Exec(`UPDATE tasks SET firstChild=?, childCount=childCount+1, descendantCount=descendantCount+1 WHERE host=? AND project=? AND id=?`, t.ID, args.Host, args.Project, t.Parent)
					PanicOn(err)
				}
				// insert new task
				_, err := tx.Exec(Strf(`INSERT INTO tasks (host, project, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, sql_task_columns), args.Host, args.Project, t.ID, t.Parent, t.FirstChild, t.NextSibling, t.User, t.Name, t.Description, t.CreatedBy, t.CreatedOn, t.MinimumTime, t.EstimatedTime, t.LoggedTime, t.EstimatedSubTime, t.LoggedSubTime, t.EstimatedExpense, t.LoggedExpense, t.EstimatedSubExpense, t.LoggedSubExpense, t.FileCount, t.FileSize, t.FileSubCount, t.FileSubSize, t.ChildCount, t.DescendantCount, t.IsParallel)
				PanicOn(err)
				// at this point the tree structure has been updated so all tasks are pointing to the correct new positions
				// all that remains to do is update aggregate values
				setAncestralChainAggregateValuesFromTask(tlbx, tx, args.Host, args.Project, args.Parent)
				tx.Commit()
				return t
			},
		},
		{
			Description:  "Update a task",
			Path:         (&task.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Update{
					Host:             app.ExampleID(),
					Project:          app.ExampleID(),
					ID:               app.ExampleID(),
					Parent:           &field.ID{V: app.ExampleID()},
					PreviousSibling:  &field.IDPtr{V: ptr.ID(app.ExampleID())},
					Name:             &field.String{V: "new name"},
					Description:      &field.StringPtr{V: ptr.String("new description")},
					IsParallel:       &field.Bool{V: true},
					User:             &field.IDPtr{V: ptr.ID(app.ExampleID())},
					EstimatedTime:    &field.UInt64{V: 123},
					EstimatedExpense: &field.UInt64{V: 123},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleTask
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				// args := a.(*task.Update)
				// me := me.Get(tlbx)
				// todo

				// return t
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
		FileSubCount:        100,
		FileSubSize:         100,
		ChildCount:          100,
		DescendantCount:     100,
		IsParallel:          true,
	}
)

func setAncestralChainAggregateValuesFromTask(tlbx app.Tlbx, tx service.Tx, host, project, task ID) {
	ancestorChain := make(IDs, 0, 20)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			i := ID{}
			PanicOn(rows.Scan(&i))
			ancestorChain = append(ancestorChain, i)
		}
	}, `CALL setAncestralChainAggregateValuesFromTask(?, ?, ?)`, host, project, task))
}

func getAncestors(tlbx app.Tlbx, tx service.Tx, host, project, ofTask ID, after *ID, limit int) []*task.Task {
	ancestors := make([]*task.Task, 0, 20)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			t, err := scan(rows)
			PanicOn(err)
			ancestors = append(ancestors, t)
		}
	}, Strf(`%s SELECT %s FROM tasks t JOIN ancestors a ON t.id = a.id WHERE t.host=? AND t.project=? AND a.n <> 0 ORDER BY a.n ASC`, sql_ancestors_cte, sql_task_columns_prefixed), host, project, ofTask, host, project, host, project))
	return ancestors
}

func getOne(tlbx app.Tlbx, tx service.Tx, host, project, one ID) *task.Task {
	row := tx.QueryRow(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.project=? AND t.id=?`, sql_task_columns_prefixed), host, project, one)
	t, err := scan(row)
	sql.PanicIfIsntNoRows(err)
	app.ReturnIf(t.ID.IsZero(), http.StatusNotFound, "")
	return t
}

type taskScanner interface {
	Scan(dest ...interface{}) error
}

func scan(ts taskScanner) (*task.Task, error) {
	t := &task.Task{}
	err := ts.Scan(
		&t.ID,
		&t.Parent,
		&t.FirstChild,
		&t.NextSibling,
		&t.User,
		&t.Name,
		&t.Description,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.MinimumTime,
		&t.EstimatedTime,
		&t.LoggedTime,
		&t.EstimatedSubTime,
		&t.LoggedSubTime,
		&t.EstimatedExpense,
		&t.LoggedExpense,
		&t.EstimatedSubExpense,
		&t.LoggedSubExpense,
		&t.FileCount,
		&t.FileSize,
		&t.FileSubCount,
		&t.FileSubSize,
		&t.ChildCount,
		&t.DescendantCount,
		&t.IsParallel)
	return t, err
}

var (
	sql_task_columns_prefixed = `t.id, t.parent, t.firstChild, t.nextSibling, t.user, t.name, t.description, t.createdBy, t.createdOn, t.minimumTime, t.estimatedTime, t.loggedTime, t.estimatedSubTime, t.loggedSubTime, t.estimatedExpense, t.loggedExpense, t.estimatedSubExpense, t.loggedSubExpense, t.fileCount, t.fileSize, t.fileSubCount, t.fileSubSize, t.childCount, t.descendantCount, t.isParallel`
	sql_task_columns          = strings.ReplaceAll(sql_task_columns_prefixed, `t.`, ``)
	sql_ancestors_cte         = `WITH RECURSIVE ancestors (n, id, parent) AS (SELECT 0, id, parent FROM tasks WHERE host=? AND project=? AND id=? UNION SELECT a.n + 1, t.id, t.parent FROM tasks t, ancestors a WHERE t.host=? AND t.project=? AND t.id = a.parent) CYCLE id RESTRICT`
)
