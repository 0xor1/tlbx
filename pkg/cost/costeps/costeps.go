package costeps

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
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/0xor1/trees/pkg/cnsts"
	"github.com/0xor1/trees/pkg/cost"
	"github.com/0xor1/trees/pkg/epsutil"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new cost",
			Path:         (&cost.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &cost.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &cost.Create{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					Value:   60,
					Note:    "I did an hours work",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleCost
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*cost.Create)
				me := me.Get(tlbx)
				app.ReturnIf(args.Value == 0, http.StatusBadRequest, "cost must be between 1 and 1440")
				args.Note = StrTrimWS(args.Note)
				validate.Str("note", args.Note, tlbx, 0, noteMaxLen)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				epsutil.MustLockProject(tx, args.Host, args.Project)
				epsutil.TaskMustExist(tx, args.Host, args.Project, args.Task)
				e := &cost.Cost{
					Task:      args.Task,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: NowMilli(),
					Value:     args.Value,
					Note:      args.Note,
				}
				// insert new cost
				_, err := tx.Exec(`INSERT INTO costs (host, project, task, id, createdBy, createdOn, value, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, args.Task, e.ID, e.CreatedBy, e.CreatedOn, e.Value, e.Note)
				PanicOn(err)
				// update task costInc
				_, err = tx.Exec(`UPDATE tasks SET costInc=costInc+? WHERE host=? AND project=? AND id=?`, args.Value, args.Host, args.Project, args.Task)
				PanicOn(err)
				// propogate aggregate costs upwards
				epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, e.ID, cnsts.TypeCost, cnsts.ActionCreated, nil, args)
				tx.Commit()
				return e
			},
		},
		{
			Description:  "Update a cost",
			Path:         (&cost.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &cost.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &cost.Update{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					ID:      app.ExampleID(),
					Value:   &field.UInt64{V: 60},
					Note:    &field.String{V: "woo"},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleCost
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*cost.Update)
				me := me.Get(tlbx)
				if args.Value == nil &&
					args.Note == nil {
					// nothing to update
					return nil
				}
				app.ReturnIf(args.Value != nil && (args.Value.V == 0), http.StatusBadRequest, "value must be 1 or greater")
				if args.Note != nil {
					args.Note.V = StrTrimWS(args.Note.V)
					validate.Str("note", args.Note.V, tlbx, 0, noteMaxLen)
				}
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				t := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "cost entry not found")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only edit your own cost entries within an hour of creating it")
				treeUpdate := false
				var diff uint64
				var sign string
				if args.Value != nil && args.Value.V != t.Value {
					if args.Value.V > t.Value {
						diff = args.Value.V - t.Value
						sign = "+"
					} else {
						diff = t.Value - args.Value.V
						sign = "-"
					}
					t.Value = args.Value.V
					treeUpdate = true
				}
				if args.Note != nil && args.Note.V != t.Note {
					t.Note = args.Note.V
				}
				_, err := tx.Exec(`UPDATE costs SET value=?, note=? WHERE host=? AND project=? AND task=? AND id=?`, t.Value, t.Note, args.Host, args.Project, t.Task, t.ID)
				PanicOn(err)
				if treeUpdate {
					_, err = tx.Exec(Strf(`UPDATE tasks SET costInc=costInc%s? WHERE host=? AND project=? AND id=?`, sign), diff, args.Host, args.Project, t.Task)
					PanicOn(err)
					epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				}
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, args.ID, cnsts.TypeCost, cnsts.ActionUpdated, nil, args)
				tx.Commit()
				return t
			},
		},
		{
			Description:  "Delete cost",
			Path:         (&cost.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &cost.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &cost.Delete{
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
				args := a.(*cost.Delete)
				me := me.Get(tlbx)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				epsutil.MustLockProject(tx, args.Host, args.Project)
				t := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only delete your own cost entries within an hour of creating it")
				// delete cost
				_, err := tx.Exec(`DELETE FROM costs WHERE host=? AND project=? AND task=? AND id=?`, args.Host, args.Project, args.Task, args.ID)
				PanicOn(err)
				_, err = tx.Exec(`UPDATE tasks SET costInc=costInc-? WHERE host=? AND project=? AND id=?`, t.Value, args.Host, args.Project, args.Task)
				PanicOn(err)
				epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				// set activities to deleted
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, args.ID, cnsts.TypeCost, cnsts.ActionDeleted, nil, nil)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "get costs",
			Path:         (&cost.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &cost.Get{
					Asc:   ptr.Bool(false),
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &cost.Get{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    ptr.ID(app.ExampleID()),
					IDs:     IDs{app.ExampleID()},
					Limit:   100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &cost.GetRes{
					Set:  []*cost.Cost{exampleCost},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*cost.Get)
				validate.MaxIDs(tlbx, "ids", args.IDs, 100)
				app.BadReqIf(
					args.CreatedOnMin != nil &&
						args.CreatedOnMax != nil &&
						args.CreatedOnMin.After(*args.CreatedOnMax),
					"createdOnMin must be before createdOnMax")
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sqlh.Limit100(args.Limit)
				res := &cost.GetRes{
					Set:  make([]*cost.Cost, 0, args.Limit),
					More: false,
				}
				qry := bytes.NewBufferString(`SELECT task, id, createdBy, createdOn, value, note FROM costs WHERE host=? AND project=?`)
				qryArgs := make([]interface{}, 0, len(args.IDs)+8)
				qryArgs = append(qryArgs, args.Host, args.Project)
				if len(args.IDs) > 0 {
					qry.WriteString(sqlh.InCondition(true, `id`, len(args.IDs)))
					qry.WriteString(sqlh.OrderByField(`id`, len(args.IDs)))
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
						qry.WriteString(Strf(` AND createdOn %s= (SELECT t.createdOn FROM costs t WHERE t.host=? AND t.project=? AND t.id=?) AND id <> ?`, sqlh.GtLtSymbol(*args.Asc)))
						qryArgs = append(qryArgs, args.Host, args.Project, *args.After, *args.After)
					}
					qry.WriteString(sqlh.OrderLimit100(`createdOn`, *args.Asc, args.Limit))
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

	noteMaxLen  = 250
	exampleCost = &cost.Cost{
		Task:      app.ExampleID(),
		ID:        app.ExampleID(),
		CreatedBy: app.ExampleID(),
		CreatedOn: app.ExampleTime(),
		Value:     60,
		Note:      "I bought something",
	}
)

func getOne(tx sql.Tx, host, project, task, id ID) *cost.Cost {
	row := tx.QueryRow(`SELECT task, id, createdBy, createdOn, value, note FROM costs WHERE host=? AND project=? AND task=? AND id=?`, host, project, task, id)
	t, err := Scan(row)
	sqlh.PanicIfIsntNoRows(err)
	return t
}

func Scan(r isql.Row) (*cost.Cost, error) {
	t := &cost.Cost{}
	err := r.Scan(
		&t.Task,
		&t.ID,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.Value,
		&t.Note)
	if t.ID.IsZero() {
		t = nil
	}
	return t, err
}
