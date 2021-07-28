package commenteps

import (
	"bytes"
	"net/http"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/comment"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
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
				me := me.AuthedGet(tlbx)
				args.Body = StrTrimWS(args.Body)
				validate.Str(tlbx, "body", args.Body, bodyMinLen, bodyMaxLen)
				tx := service.Get(tlbx).Data().BeginWrite()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleWriter)
				epsutil.TaskMustExist(tx, args.Host, args.Project, args.Task)
				c := &comment.Comment{
					Task:      args.Task,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: tlbx.Start(),
					Body:      args.Body,
				}
				// insert new comment
				_, err := tx.Exec(`INSERT INTO comments (host, project, task, id, createdBy, createdOn, body) VALUES (?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, args.Task, c.ID, c.CreatedBy, c.CreatedOn, c.Body)
				PanicOn(err)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Task, c.ID, cnsts.TypeComment, cnsts.ActionCreated, nil, StrEllipsis(args.Body, 50), StrEllipsis(args.Body, 1000), nil)
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
					Body:    "woo comment changed",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleComment
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*comment.Update)
				me := me.AuthedGet(tlbx)
				args.Body = StrTrimWS(args.Body)
				validate.Str(tlbx, "body", args.Body, bodyMinLen, bodyMaxLen)
				tx := service.Get(tlbx).Data().BeginWrite()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				c := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(c == nil, http.StatusNotFound, "comment entry not found")
				app.ReturnIf(role == cnsts.RoleWriter && !c.CreatedBy.Equal(me), http.StatusForbidden, "you may only update your own comment")
				c.Body = args.Body
				_, err := tx.Exec(`UPDATE comments SET body=? WHERE host=? AND project=? AND task=? AND id=?`, c.Body, args.Host, args.Project, c.Task, c.ID)
				PanicOn(err)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Task, args.ID, cnsts.TypeComment, cnsts.ActionUpdated, nil, StrEllipsis(args.Body, 50), StrEllipsis(args.Body, 1000), nil)
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
				me := me.AuthedGet(tlbx)
				tx := service.Get(tlbx).Data().BeginWrite()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				epsutil.MustLockProject(tx, args.Host, args.Project)
				c := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(c == nil, http.StatusNotFound, "")
				app.ReturnIf(role == cnsts.RoleWriter && !c.CreatedBy.Equal(me), http.StatusForbidden, "you may only delete your own comment")
				// delete comment
				_, err := tx.Exec(`DELETE FROM comments WHERE host=? AND project=? AND task=? AND id=?`, args.Host, args.Project, args.Task, args.ID)
				PanicOn(err)
				// set activities to deleted
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Task, args.ID, cnsts.TypeComment, cnsts.ActionDeleted, nil, StrEllipsis(c.Body, 50), StrEllipsis(c.Body, 1000), nil)
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
				tx := service.Get(tlbx).Data().BeginRead()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sqlh.Limit100(args.Limit)
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
				qry.WriteString(sqlh.OrderLimit100(`createdOn`, false, args.Limit))
				PanicOn(tx.Query(func(rows isql.Rows) {
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
				tx.Commit()
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

func getOne(tx sql.Tx, host, project, task, id ID) *comment.Comment {
	row := tx.QueryRow(`SELECT task, id, createdBy, createdOn, body FROM comments WHERE host=? AND project=? AND task=? AND id=?`, host, project, task, id)
	t, err := Scan(row)
	sqlh.PanicIfIsntNoRows(err)
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
