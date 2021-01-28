package fileeps

import (
	"bytes"
	"net/http"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/file"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/0xor1/trees/pkg/epsutil"
	"github.com/0xor1/trees/pkg/task/taskeps"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Put a file",
			Path:         (&file.Put{}).Path(),
			Timeout:      300000, // 5 mins
			MaxBodyBytes: 5 * app.GB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &app.UpStream{
					Args: &file.PutArgs{},
				}
			},
			GetExampleArgs: func() interface{} {
				res := &app.UpStream{}
				res.Name = "my_cool_file.pdf"
				res.Size = 100
				res.Type = "application/pdf"
				return res
			},
			GetExampleResponse: func() interface{} {
				return &file.PutRes{
					Task: taskeps.ExampleTask,
					File: exampleFile,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*app.UpStream)
				defer args.Content.Close()
				me := me.Get(tlbx)
				innerArgs := args.Args.(*file.PutArgs)
				app.BadReqIf(innerArgs.Host.IsZero() || innerArgs.Project.IsZero() || innerArgs.Task.IsZero(), "Content-Args header must be set")
				app.ReturnIf(args.Size > maxFileSize, http.StatusBadRequest, "max file size is %d", maxFileSize)
				args.Name = StrTrimWS(args.Name)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				epsutil.IMustHaveAccess(tlbx, innerArgs.Host, innerArgs.Project, cnsts.RoleWriter)
				f := &file.File{
					Task:      innerArgs.Task,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: NowMilli(),
					Size:      uint64(args.Size),
					Type:      args.Type,
					Name:      args.Name,
				}
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				// insert new file
				_, err := tx.Exec(`INSERT INTO files (host, project, task, id, name, createdBy, createdOn, size, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, innerArgs.Host, innerArgs.Project, innerArgs.Task, f.ID, f.Name, f.CreatedBy, f.CreatedOn, f.Size, f.Type)
				PanicOn(err)
				// update task file values
				_, err = tx.Exec(`UPDATE tasks SET fileN=fileN+1, fileSize=fileSize+? WHERE host=? AND project=? AND id=?`, f.Size, innerArgs.Host, innerArgs.Project, innerArgs.Task)
				PanicOn(err)
				// propogate aggregate sizes upwards
				epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, innerArgs.Host, innerArgs.Project, innerArgs.Task)
				epsutil.LogActivity(tlbx, tx, innerArgs.Host, innerArgs.Project, &innerArgs.Task, f.ID, cnsts.TypeFile, cnsts.ActionCreated, &f.Name, struct {
					Name string `json:"name"`
					Size uint64 `json:"size"`
					Type string `json:"type"`
				}{
					Name: args.Name,
					Size: uint64(args.Size),
					Type: args.Type,
				})
				srv.Store().MustStreamUp(cnsts.FileBucket, store.Key("", innerArgs.Host, innerArgs.Project, innerArgs.Task, f.ID), f.Name, f.Type, int64(f.Size), false, true, 5*time.Minute, args.Content)
				res := &file.PutRes{
					Task: taskeps.GetOne(tx, innerArgs.Host, innerArgs.Project, innerArgs.Task),
					File: f,
				}
				tx.Commit()
				return res
			},
		},
		{
			Description:  "get file content",
			Path:         (&file.GetContent{}).Path(),
			Timeout:      300000, // 5 mins
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &file.GetContent{}
			},
			GetExampleArgs: func() interface{} {
				return &file.GetContent{
					Host:       app.ExampleID(),
					Project:    app.ExampleID(),
					Task:       app.ExampleID(),
					ID:         app.ExampleID(),
					IsDownload: true,
				}
			},
			GetExampleResponse: func() interface{} {
				return &app.DownStream{}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*file.GetContent)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				f := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				tx.Commit()
				app.ReturnIf(f == nil, http.StatusNotFound, "file not found")
				res := &app.DownStream{
					ID:         f.ID,
					IsDownload: args.IsDownload,
				}
				res.Name = f.Name
				res.Size = int64(f.Size)
				res.Type = f.Type
				_, _, _, res.Content = srv.Store().MustGet(cnsts.FileBucket, store.Key("", args.Host, args.Project, args.Task, args.ID))
				return res
			},
		},
		{
			Description:  "get files",
			Path:         (&file.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &file.Get{
					Asc:   ptr.Bool(false),
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &file.Get{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    ptr.ID(app.ExampleID()),
					IDs:     IDs{app.ExampleID()},
					Limit:   100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &file.GetRes{
					Set:  []*file.File{exampleFile},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*file.Get)
				validate.MaxIDs(tlbx, "ids", args.IDs, 100)
				app.BadReqIf(
					args.CreatedOnMin != nil &&
						args.CreatedOnMax != nil &&
						args.CreatedOnMin.After(*args.CreatedOnMax),
					"createdOnMin must be before createdOnMax")
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sqlh.Limit100(args.Limit)
				res := &file.GetRes{
					Set:  make([]*file.File, 0, args.Limit),
					More: false,
				}
				qry := bytes.NewBufferString(`SELECT task, id, name, createdBy, createdOn, size, type FROM files WHERE host=? AND project=?`)
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
						qry.WriteString(Strf(` AND createdOn %s= (SELECT f.createdOn FROM files f WHERE f.host=? AND f.project=? AND f.id=?) AND id <> ?`, sqlh.GtLtSymbol(*args.Asc)))
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
		{
			Description:  "Delete file",
			Path:         (&file.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &file.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &file.Delete{
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
				args := a.(*file.Delete)
				me := me.Get(tlbx)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "")
				epsutil.MustLockProject(tx, args.Host, args.Project)
				f := getOne(tx, args.Host, args.Project, args.Task, args.ID)
				app.ReturnIf(f == nil, http.StatusNotFound, "file not found")
				app.ReturnIf(role == cnsts.RoleWriter && (!f.CreatedBy.Equal(me) || f.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only delete your own file entries within an hour of creating it")
				// delete file
				_, err := tx.Exec(`DELETE FROM files WHERE host=? AND project=? AND task=? AND id=?`, args.Host, args.Project, args.Task, args.ID)
				PanicOn(err)
				_, err = tx.Exec(`UPDATE tasks SET fileN=fileN-1, fileSize=fileSize-? WHERE host=? AND project=? AND id=?`, f.Size, args.Host, args.Project, args.Task)
				PanicOn(err)
				epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				// set activities to deleted
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, args.ID, cnsts.TypeFile, cnsts.ActionDeleted, &f.Name, nil)
				tx.Commit()
				return nil
			},
		},
	}

	nameMinLen  = 1
	nameMaxLen  = 250
	maxFileSize = 5 * app.GB
	exampleFile = &file.File{
		Task:      app.ExampleID(),
		ID:        app.ExampleID(),
		CreatedBy: app.ExampleID(),
		CreatedOn: app.ExampleTime(),
		Size:      60,
		Type:      "application/pdf",
		Name:      "my_file.pdf",
	}
)

func getOne(tx sql.Tx, host, project, task, id ID) *file.File {
	row := tx.QueryRow(`SELECT task, id, name, createdBy, createdOn, size, type FROM files WHERE host=? AND project=? AND task=? AND id=?`, host, project, task, id)
	t, err := Scan(row)
	sqlh.PanicIfIsntNoRows(err)
	return t
}

func Scan(r isql.Row) (*file.File, error) {
	f := &file.File{}
	err := r.Scan(
		&f.Task,
		&f.ID,
		&f.Name,
		&f.CreatedBy,
		&f.CreatedOn,
		&f.Size,
		&f.Type)
	if f.ID.IsZero() {
		f = nil
	}
	return f, err
}
