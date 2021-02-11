package vitemeps

import (
	"bytes"
	"net/http"
	time_ "time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/vitem"
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
	"github.com/0xor1/trees/pkg/epsutil"
	"github.com/0xor1/trees/pkg/task/taskeps"
)

type extraInfo struct {
	Type vitem.Type `json:"type"`
	Note string     `json:"note"`
	Est  *uint64    `json:"est,omitempty"`
	Inc  uint64     `json:"inc"`
}

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new vitem",
			Path:         (&vitem.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &vitem.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &vitem.Create{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					Type:    vitem.TypeTime,
					Est:     ptr.Uint64(35),
					Inc:     60,
					Note:    "I did an hours work",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleVitemRes
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*vitem.Create)
				args.Type.Validate()
				me := me.Get(tlbx)
				switch args.Type {
				case vitem.TypeTime:
					app.ReturnIf(args.Inc == 0 || args.Inc > timeValueMax, http.StatusBadRequest, "time inc must be between 1 and 1440")
				case vitem.TypeCost:
					app.ReturnIf(args.Inc == 0, http.StatusBadRequest, "cost inc must be > 1")
					// default:
					//  // uneeded check due to args.Type.Validate()
					// 	app.BadReqIf(true, "unknown type value %s", args.Type)
				}
				args.Note = StrTrimWS(args.Note)
				validate.Str("note", args.Note, tlbx, 0, noteMaxLen)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleWriter)
				epsutil.MustLockProject(tx, args.Host, args.Project)
				tsk := taskeps.GetOne(tx, args.Host, args.Project, args.Task)
				app.ReturnIf(tsk == nil, http.StatusNotFound, "task not found")
				gtr := func(est bool) uint64 {
					return taskValGetter(tsk, args.Type, est)()
				}
				if args.Est != nil {
					// set the tasks estimated type value
					taskValSetter(tsk, args.Type, true)(*args.Est)
				}
				// add to teh tasks incurred type value
				taskValSetter(tsk, args.Type, false)(gtr(false) + args.Inc)
				i := &vitem.Vitem{
					Task:      args.Task,
					Type:      args.Type,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: tlbx.Start(),
					Inc:       args.Inc,
					Note:      args.Note,
				}
				// insert new vitem
				_, err := tx.Exec(`INSERT INTO vitems (host, project, task, type, id, createdBy, createdOn, inc, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, args.Task, args.Type, i.ID, i.CreatedBy, i.CreatedOn, i.Inc, i.Note)
				PanicOn(err)
				// update task vals
				_, err = tx.Exec(Strf(`UPDATE tasks SET %sEst=?, %sInc=? WHERE host=? AND project=? AND id=?`, i.Type, i.Type), gtr(true), gtr(false), args.Host, args.Project, args.Task)
				PanicOn(err)
				// propogate aggregate values upwards
				ancestors := epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Task, i.ID, cnsts.TypeVitem, cnsts.ActionCreated, nil, &extraInfo{
					Type: args.Type,
					Note: StrEllipsis(args.Note, 50),
					Est:  args.Est,
					Inc:  args.Inc,
				}, &extraInfo{
					Type: args.Type,
					Note: args.Note,
					Est:  args.Est,
					Inc:  args.Inc,
				}, ancestors)
				tx.Commit()
				return vitem.VitemRes{
					Task: tsk,
					Item: i,
				}
			},
		},
		{
			Description:  "Update a vitem",
			Path:         (&vitem.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &vitem.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &vitem.Update{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					Type:    vitem.TypeTime,
					ID:      app.ExampleID(),
					Inc:     &field.UInt64{V: 60},
					Note:    &field.String{V: "woo"},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleVitemRes
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*vitem.Update)
				args.Type.Validate()
				me := me.Get(tlbx)
				if args.Inc == nil &&
					args.Note == nil {
					// nothing to update
					return nil
				}
				switch args.Type {
				case vitem.TypeTime:
					app.ReturnIf(args.Inc != nil && (args.Inc.V == 0 || args.Inc.V > timeValueMax), http.StatusBadRequest, "inc must be between 1 and 1440")
				case vitem.TypeCost:
					app.ReturnIf(args.Inc != nil && args.Inc.V == 0, http.StatusBadRequest, "inc must be > 1")
					// default:
					//  // uneeded check due to args.Type.Validate()
					// 	app.BadReqIf(true, "unknown type value %s", args.Type)
				}
				if args.Note != nil {
					args.Note.V = StrTrimWS(args.Note.V)
					validate.Str("note", args.Note.V, tlbx, 0, noteMaxLen)
				}
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				t := getOne(tx, args.Host, args.Project, args.Task, args.ID, args.Type)
				app.ReturnIf(t == nil, http.StatusNotFound, "vitem entry not found")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.CreatedOn.Before(Now().Add(-1*time_.Hour))), http.StatusForbidden, "you may only edit your own vitem entry within an hour of creating it")
				treeUpdate := false
				var diff uint64
				var sign string
				isChange := false
				if args.Inc != nil && args.Inc.V != t.Inc {
					isChange = true
					if args.Inc.V > t.Inc {
						diff = args.Inc.V - t.Inc
						sign = "+"
					} else {
						diff = t.Inc - args.Inc.V
						sign = "-"
					}
					t.Inc = args.Inc.V
					treeUpdate = true
				}
				if args.Note != nil && args.Note.V != t.Note {
					isChange = true
					t.Note = args.Note.V
				}
				if !isChange {
					return nil
				}
				_, err := tx.Exec(`UPDATE vitems SET inc=?, note=? WHERE host=? AND project=? AND task=? AND id=? AND type=? `, t.Inc, t.Note, args.Host, args.Project, t.Task, t.ID, t.Type)
				PanicOn(err)
				var ancestors IDs
				if treeUpdate {
					_, err = tx.Exec(Strf(`UPDATE tasks SET %sInc=%sInc%s? WHERE host=? AND project=? AND id=?`, args.Type, args.Type, sign), diff, args.Host, args.Project, t.Task)
					PanicOn(err)
					ancestors = epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				}
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Task, args.ID, cnsts.TypeVitem, cnsts.ActionUpdated, nil, &extraInfo{
					Type: t.Type,
					Note: StrEllipsis(t.Note, 50),
					Inc:  t.Inc,
				}, &extraInfo{
					Type: t.Type,
					Note: t.Note,
					Inc:  t.Inc,
				}, ancestors)
				tsk := taskeps.GetOne(tx, args.Host, args.Project, args.Task)
				tx.Commit()
				return &vitem.VitemRes{
					Task: tsk,
					Item: t,
				}
			},
		},
		{
			Description:  "Delete time",
			Path:         (&vitem.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &vitem.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &vitem.Delete{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					ID:      app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return taskeps.ExampleTask
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*vitem.Delete)
				args.Type.Validate()
				me := me.Get(tlbx)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				epsutil.MustLockProject(tx, args.Host, args.Project)
				v := getOne(tx, args.Host, args.Project, args.Task, args.ID, args.Type)
				app.ReturnIf(v == nil, http.StatusNotFound, "vitem not found")
				app.ReturnIf(role == cnsts.RoleWriter && (!v.CreatedBy.Equal(me) || v.CreatedOn.Before(Now().Add(-1*time_.Hour))), http.StatusForbidden, "you may only delete your own vitem entry within an hour of creating it")
				// delete time
				_, err := tx.Exec(`DELETE FROM vitems WHERE host=? AND project=? AND task=? AND id=? AND type=?`, args.Host, args.Project, args.Task, args.ID, args.Type)
				PanicOn(err)
				_, err = tx.Exec(Strf(`UPDATE tasks SET %sInc=%sInc-? WHERE host=? AND project=? AND id=?`, args.Type, args.Type), v.Inc, args.Host, args.Project, args.Task)
				PanicOn(err)
				ancestors := epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				// set activities to deleted
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Task, args.ID, cnsts.TypeVitem, cnsts.ActionDeleted, nil, &extraInfo{
					Type: v.Type,
					Note: StrEllipsis(v.Note, 50),
					Inc:  v.Inc,
				}, &extraInfo{
					Type: v.Type,
					Note: v.Note,
					Inc:  v.Inc,
				}, ancestors)
				tsk := taskeps.GetOne(tx, args.Host, args.Project, args.Task)
				tx.Commit()
				return tsk
			},
		},
		{
			Description:  "get vitems",
			Path:         (&vitem.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &vitem.Get{
					Asc:   ptr.Bool(false),
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &vitem.Get{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    ptr.ID(app.ExampleID()),
					IDs:     IDs{app.ExampleID()},
					Limit:   100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &vitem.GetRes{
					Set:  []*vitem.Vitem{exampleVitem},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*vitem.Get)
				args.Type.Validate()
				validate.MaxIDs(tlbx, "ids", args.IDs, 100)
				app.BadReqIf(
					args.CreatedOnMin != nil &&
						args.CreatedOnMax != nil &&
						args.CreatedOnMin.After(*args.CreatedOnMax),
					"createdOnMin must be before createdOnMax")
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sqlh.Limit100(args.Limit)
				res := &vitem.GetRes{
					Set:  make([]*vitem.Vitem, 0, args.Limit),
					More: false,
				}
				qry := bytes.NewBufferString(`SELECT task, type, id, createdBy, createdOn, inc, note FROM vitems WHERE host=? AND project=? AND type=?`)
				qryArgs := make([]interface{}, 0, len(args.IDs)+9)
				qryArgs = append(qryArgs, args.Host, args.Project, args.Type)
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
						qry.WriteString(Strf(` AND createdOn %s= (SELECT v.createdOn FROM vitems v WHERE v.host=? AND v.project=? AND v.type=? AND v.id=?) AND id <> ?`, sqlh.GtLtSymbol(*args.Asc)))
						qryArgs = append(qryArgs, args.Host, args.Project, args.Type, *args.After, *args.After)
					}
					qry.WriteString(sqlh.OrderLimit100(`createdOn`, *args.Asc, args.Limit))
				}
				PanicOn(tx.Query(func(rows isql.Rows) {
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
				tx.Commit()
				return res
			},
		},
	}

	timeValueMax    uint64 = 1440 // 24hrs in mins
	noteMaxLen             = 250
	exampleVitemRes        = &vitem.VitemRes{
		Task: taskeps.ExampleTask,
		Item: exampleVitem,
	}
	exampleVitem = &vitem.Vitem{
		Task:      app.ExampleID(),
		Type:      vitem.TypeTime,
		ID:        app.ExampleID(),
		CreatedBy: app.ExampleID(),
		CreatedOn: app.ExampleTime(),
		Inc:       60,
		Note:      "I did something",
	}
)

func taskValSetter(tsk *task.Task, typ vitem.Type, est bool) func(uint64) {
	switch typ {
	case vitem.TypeTime:
		if est {
			return func(v uint64) {
				tsk.TimeEst = v
			}
		} else {
			return func(v uint64) {
				tsk.TimeInc = v
			}
		}
	case vitem.TypeCost:
		if est {
			return func(v uint64) {
				tsk.CostEst = v
			}
		} else {
			return func(v uint64) {
				tsk.CostInc = v
			}
		}
		// default:
		//  // uneeded check due to args.Type.Validate()
		// 	app.BadReqIf(true, "unknown type value %s", typ)
		//  return nil
	}
	return nil
}

func taskValGetter(tsk *task.Task, typ vitem.Type, est bool) func() uint64 {
	switch typ {
	case vitem.TypeTime:
		if est {
			return func() uint64 {
				return tsk.TimeEst
			}
		} else {
			return func() uint64 {
				return tsk.TimeInc
			}
		}
	case vitem.TypeCost:
		if est {
			return func() uint64 {
				return tsk.CostEst
			}
		} else {
			return func() uint64 {
				return tsk.CostInc
			}
		}
		// default:
		//  // uneeded check due to args.Type.Validate()
		// 	app.BadReqIf(true, "unknown type value %s", typ)
		//  return nil
	}
	return nil
}

func getOne(tx sql.Tx, host, project, task, id ID, typ vitem.Type) *vitem.Vitem {
	row := tx.QueryRow(`SELECT task, type, id, createdBy, createdOn, inc, note FROM vitems WHERE host=? AND project=? AND task=? AND id=? AND type=?`, host, project, task, id, typ)
	t, err := Scan(row)
	sqlh.PanicIfIsntNoRows(err)
	return t
}

func Scan(r isql.Row) (*vitem.Vitem, error) {
	t := &vitem.Vitem{}
	err := r.Scan(
		&t.Task,
		&t.Type,
		&t.ID,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.Inc,
		&t.Note)
	if t.ID.IsZero() {
		t = nil
	}
	return t, err
}
