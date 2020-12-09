package commenteps

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
	"github.com/0xor1/trees/pkg/comment"
	"github.com/0xor1/trees/pkg/epsutil"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new comment",
			Path:         (&comment.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: 50 * app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &comment.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &comment.Create{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					Body:    "comment body",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleComment
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*comment.Create)
				me := me.Get(tlbx)
				args.Body = StrTrimWS(args.Body)
				validate.Str("body", args.Body, tlbx, bodyMinLen, bodyMaxLen)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				epsutil.TaskMustExist(tx, args.Host, args.Project, args.Task)
				c := &comment.Comment{
					Task:      args.Task,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: NowMilli(),
					Body:      args.Body,
				}
				// insert new comment
				_, err := tx.Exec(`INSERT INTO comments (host, project, task, id, createdBy, createdOn, body) VALUES (?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, args.Task, c.ID, c.CreatedBy, c.CreatedOn, c.Body)
				PanicOn(err)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, c.ID, cnsts.TypeComment, cnsts.ActionCreated, nil, args)
				tx.Commit()
				return c
			},
		},
		{
			Description:  "Update a comment",
			Path:         (&comment.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: 50 * app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &comment.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &comment.Update{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					ID:      app.ExampleID(),
					Body:    &field.String{V: "woo comment changed"},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleComment
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*comment.Update)
				me := me.Get(tlbx)
				if args.Body == nil {
					// nothing to update
					return nil
				}
				args.Body.V = StrTrimWS(args.Body.V)
				validate.Str("body", args.Body.V, tlbx, bodyMinLen, bodyMaxLen)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				c := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(c == nil, http.StatusNotFound, "comment entry not found")
				app.ReturnIf(role == cnsts.RoleWriter && (!c.CreatedBy.Equal(me) || c.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only edit your own comment entries within an hour of creating it")
				c.Body = args.Body.V
				_, err := tx.Exec(`UPDATE comments SET body=? WHERE host=? AND project=? AND task=? AND id=?`, c.Body, args.Host, args.Project, c.Task, c.ID)
				PanicOn(err)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, args.ID, cnsts.TypeComment, cnsts.ActionUpdated, nil, args.Body)
				tx.Commit()
				return c
			},
		},
		{
			Description:  "Delete comment",
			Path:         (&comment.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &comment.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &comment.Delete{
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
				args := a.(*comment.Delete)
				me := me.Get(tlbx)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				epsutil.MustLockProject(tx, args.Host, args.Project)
				t := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only delete your own comment entries within an hour of creating it")
				// delete comment
				_, err := tx.Exec(`DELETE FROM comments WHERE host=? AND project=? AND task=? AND id=?`, args.Host, args.Project, args.Task, args.ID)
				PanicOn(err)
				// set activities to deleted
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, args.ID, cnsts.TypeComment, cnsts.ActionDeleted, nil, nil)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "get comments",
			Path:         (&comment.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &comment.Get{
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &comment.Get{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    ptr.ID(app.ExampleID()),
					Limit:   100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &comment.GetRes{
					Set:  []*comment.Comment{exampleComment},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*comment.Get)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sql.Limit100(args.Limit)
				res := &comment.GetRes{
					Set:  make([]*comment.Comment, 0, args.Limit),
					More: false,
				}
				qry := bytes.NewBufferString(`SELECT task, id, createdBy, createdOn, body FROM comments WHERE host=? AND project=?`)
				qryArgs := make([]interface{}, 0, 8)
				qryArgs = append(qryArgs, args.Host, args.Project)
				if args.Task != nil {
					qry.WriteString(` AND task=?`)
					qryArgs = append(qryArgs, args.Task)
				}
				if args.After != nil {
					qry.WriteString(` AND createdOn <= (SELECT c.createdOn FROM comments c WHERE c.host=? AND c.project=? AND c.id=?) AND id <> ?`)
					qryArgs = append(qryArgs, args.Host, args.Project, *args.After, *args.After)
				}
				qry.WriteString(sql.OrderLimit100(`createdOn`, false, args.Limit))
				PanicOn(service.Get(tlbx).Data().Query(func(rows isql.Rows) {
					iLimit := int(args.Limit)
					for rows.Next() {
						if len(res.Set)+1 == iLimit {
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

	bodyMinLen     = 1
	bodyMaxLen     = 10000
	exampleComment = &comment.Comment{
		Task:      app.ExampleID(),
		ID:        app.ExampleID(),
		CreatedBy: app.ExampleID(),
		CreatedOn: app.ExampleTime(),
		Body:      "this is a comment body",
	}
)

func getOne(tx service.Tx, host, project, task, id ID) *comment.Comment {
	row := tx.QueryRow(`SELECT task, id, createdBy, createdOn, body FROM comments WHERE host=? AND project=? AND task=? AND id=?`, host, project, task, id)
	t, err := Scan(row)
	sql.PanicIfIsntNoRows(err)
	return t
}

func Scan(r isql.Row) (*comment.Comment, error) {
	t := &comment.Comment{}
	err := r.Scan(
		&t.Task,
		&t.ID,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.Body)
	if t.ID.IsZero() {
		t = nil
	}
	return t, err
}
