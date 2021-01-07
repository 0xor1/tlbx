package expenseeps

import (
	"bytes"
	"net/http"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/0xor1/trees/pkg/cnsts"
	"github.com/0xor1/trees/pkg/epsutil"
	"github.com/0xor1/trees/pkg/expense"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new expense",
			Path:         (&expense.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &expense.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &expense.Create{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					Cost:    60,
					Note:    "I did an hours work",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleExpense
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*expense.Create)
				me := me.Get(tlbx)
				app.ReturnIf(args.Cost == 0, http.StatusBadRequest, "cost must be between 1 and 1440")
				args.Note = StrTrimWS(args.Note)
				validate.Str("note", args.Note, tlbx, 0, noteMaxLen)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				epsutil.MustLockProject(tx, args.Host, args.Project)
				epsutil.TaskMustExist(tx, args.Host, args.Project, args.Task)
				e := &expense.Expense{
					Task:      args.Task,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: NowMilli(),
					Cost:      args.Cost,
					Note:      args.Note,
				}
				// insert new expense
				_, err := tx.Exec(`INSERT INTO expenses (host, project, task, id, createdBy, createdOn, cost, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, args.Task, e.ID, e.CreatedBy, e.CreatedOn, e.Cost, e.Note)
				PanicOn(err)
				// update task costInc
				_, err = tx.Exec(`UPDATE tasks SET costInc=costInc+? WHERE host=? AND project=? AND id=?`, args.Cost, args.Host, args.Project, args.Task)
				PanicOn(err)
				// propogate aggregate costs upwards
				epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, e.ID, cnsts.TypeExpense, cnsts.ActionCreated, nil, args)
				tx.Commit()
				return e
			},
		},
		{
			Description:  "Update a expense",
			Path:         (&expense.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &expense.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &expense.Update{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					ID:      app.ExampleID(),
					Cost:    &field.UInt64{V: 60},
					Note:    &field.String{V: "woo"},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleExpense
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*expense.Update)
				me := me.Get(tlbx)
				if args.Cost == nil &&
					args.Note == nil {
					// nothing to update
					return nil
				}
				app.ReturnIf(args.Cost != nil && (args.Cost.V == 0), http.StatusBadRequest, "cost must be 1 or greater")
				if args.Note != nil {
					args.Note.V = StrTrimWS(args.Note.V)
					validate.Str("note", args.Note.V, tlbx, 0, noteMaxLen)
				}
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				t := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "expense entry not found")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only edit your own expense entries within an hour of creating it")
				treeUpdate := false
				var diff uint64
				var sign string
				if args.Cost != nil && args.Cost.V != t.Cost {
					if args.Cost.V > t.Cost {
						diff = args.Cost.V - t.Cost
						sign = "+"
					} else {
						diff = t.Cost - args.Cost.V
						sign = "-"
					}
					t.Cost = args.Cost.V
					treeUpdate = true
				}
				if args.Note != nil && args.Note.V != t.Note {
					t.Note = args.Note.V
				}
				_, err := tx.Exec(`UPDATE expenses SET cost=?, note=? WHERE host=? AND project=? AND task=? AND id=?`, t.Cost, t.Note, args.Host, args.Project, t.Task, t.ID)
				PanicOn(err)
				if treeUpdate {
					_, err = tx.Exec(Strf(`UPDATE tasks SET costInc=costInc%s? WHERE host=? AND project=? AND id=?`, sign), diff, args.Host, args.Project, t.Task)
					PanicOn(err)
					epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				}
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, args.ID, cnsts.TypeExpense, cnsts.ActionUpdated, nil, args)
				tx.Commit()
				return t
			},
		},
		{
			Description:  "Delete expense",
			Path:         (&expense.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &expense.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &expense.Delete{
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
				args := a.(*expense.Delete)
				me := me.Get(tlbx)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				epsutil.MustLockProject(tx, args.Host, args.Project)
				t := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only delete your own expense entries within an hour of creating it")
				// delete expense
				_, err := tx.Exec(`DELETE FROM expenses WHERE host=? AND project=? AND task=? AND id=?`, args.Host, args.Project, args.Task, args.ID)
				PanicOn(err)
				_, err = tx.Exec(`UPDATE tasks SET costInc=costInc-? WHERE host=? AND project=? AND id=?`, t.Cost, args.Host, args.Project, args.Task)
				PanicOn(err)
				epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				// set activities to deleted
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, args.ID, cnsts.TypeExpense, cnsts.ActionDeleted, nil, nil)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "get expenses",
			Path:         (&expense.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &expense.Get{
					Asc:   ptr.Bool(false),
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &expense.Get{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    ptr.ID(app.ExampleID()),
					IDs:     IDs{app.ExampleID()},
					Limit:   100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &expense.GetRes{
					Set:  []*expense.Expense{exampleExpense},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*expense.Get)
				validate.MaxIDs(tlbx, "ids", args.IDs, 100)
				app.BadReqIf(
					args.CreatedOnMin != nil &&
						args.CreatedOnMax != nil &&
						args.CreatedOnMin.After(*args.CreatedOnMax),
					"createdOnMin must be before createdOnMax")
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sql.Limit100(args.Limit)
				res := &expense.GetRes{
					Set:  make([]*expense.Expense, 0, args.Limit),
					More: false,
				}
				qry := bytes.NewBufferString(`SELECT task, id, createdBy, createdOn, cost, note FROM expenses WHERE host=? AND project=?`)
				qryArgs := make([]interface{}, 0, len(args.IDs)+8)
				qryArgs = append(qryArgs, args.Host, args.Project)
				if len(args.IDs) > 0 {
					qry.WriteString(sql.InCondition(true, `id`, len(args.IDs)))
					qry.WriteString(sql.OrderByField(`id`, len(args.IDs)))
					ids := args.IDs.ToIs()
					qryArgs = append(qryArgs, ids...)
					qryArgs = append(qryArgs, ids...)
				} else {
					if args.Task != nil {
						qry.WriteString(` AND task=?`)
						qryArgs = append(qryArgs, args.Task)
					}
					if args.CreatedOnMin != nil {
						qry.WriteString(` AND createdOn>=?`)
						qryArgs = append(qryArgs, *args.CreatedOnMin)
					}
					if args.CreatedOnMax != nil {
						qry.WriteString(` AND createdOn<=?`)
						qryArgs = append(qryArgs, *args.CreatedOnMax)
					}
					if args.CreatedBy != nil {
						qry.WriteString(` AND createdBy=?`)
						qryArgs = append(qryArgs, *args.CreatedBy)
					}
					if args.After != nil {
						qry.WriteString(Strf(` AND createdOn %s= (SELECT t.createdOn FROM expenses t WHERE t.host=? AND t.project=? AND t.id=?) AND id <> ?`, sql.GtLtSymbol(*args.Asc)))
						qryArgs = append(qryArgs, args.Host, args.Project, *args.After, *args.After)
					}
					qry.WriteString(sql.OrderLimit100(`createdOn`, *args.Asc, args.Limit))
				}
				PanicOn(service.Get(tlbx).Data().Query(func(rows isql.Rows) {
					iLimit := int(args.Limit)
					for rows.Next() {
						if len(args.IDs) == 0 && len(res.Set)+1 == iLimit {
							res.More = true
							break
						}
						t, err := Scan(rows)
						PanicOn(err)
						res.Set = append(res.Set, t)
					}
				}, qry.String(), qryArgs...))
				return res
			},
		},
	}

	noteMaxLen     = 250
	exampleExpense = &expense.Expense{
		Task:      app.ExampleID(),
		ID:        app.ExampleID(),
		CreatedBy: app.ExampleID(),
		CreatedOn: app.ExampleTime(),
		Cost:      60,
		Note:      "I bought something",
	}
)

func getOne(tx service.Tx, host, project, task, id ID) *expense.Expense {
	row := tx.QueryRow(`SELECT task, id, createdBy, createdOn, cost, note FROM expenses WHERE host=? AND project=? AND task=? AND id=?`, host, project, task, id)
	t, err := Scan(row)
	sql.PanicIfIsntNoRows(err)
	return t
}

func Scan(r isql.Row) (*expense.Expense, error) {
	t := &expense.Expense{}
	err := r.Scan(
		&t.Task,
		&t.ID,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.Cost,
		&t.Note)
	if t.ID.IsZero() {
		t = nil
	}
	return t, err
}
