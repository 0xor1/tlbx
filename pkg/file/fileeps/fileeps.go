package fileeps

import (
	"bytes"
	"net/http"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/0xor1/trees/pkg/cnsts"
	"github.com/0xor1/trees/pkg/epsutil"
	"github.com/0xor1/trees/pkg/file"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Get a presigned put url to upload a new file",
			Path:         (&file.GetPresignedPutUrl{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &file.GetPresignedPutUrl{}
			},
			GetExampleArgs: func() interface{} {
				return &file.GetPresignedPutUrl{
					Host:     app.ExampleID(),
					Project:  app.ExampleID(),
					Task:     app.ExampleID(),
					Name:     "my_cool_file.pdf",
					MimeType: "application/pdf",
					Size:     100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &file.GetPresignedPutUrlRes{
					URL: "https://minio.com/put/your/file/here",
					ID:  app.ExampleID(),
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*file.GetPresignedPutUrl)
				me := me.Get(tlbx)
				app.ReturnIf(args.Size == 0, http.StatusBadRequest, "size must be between 1 and 1440")
				args.Name = StrTrimWS(args.Name)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				f := &file.File{
					Task:      args.Task,
					ID:        tlbx.NewID(),
					CreatedBy: me,
					CreatedOn: NowMilli(),
					Size:      args.Size,
					MimeType:  args.MimeType,
					Name:      args.Name,
				}
				// insert new file NOT finalized
				_, err := tx.Exec(`INSERT INTO files (host, project, task, isFinalized, id, name, createdBy, createdOn, size, mimeType) VALUES (?, ?, ?, 0, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, args.Task, f.ID, f.Name, f.CreatedBy, f.CreatedOn, f.Size, f.MimeType)
				PanicOn(err)
				res := &file.GetPresignedPutUrlRes{
					URL: srv.Store().MustPresignedPutUrl(cnsts.TempFileBucket, store.Key("", args.Host, args.Project, args.Task, f.ID), f.Name, f.MimeType, int64(f.Size)),
					ID:  f.ID,
				}
				tx.Commit()
				return res
			},
		},
		{
			Description:  "Finalize an uploaded file",
			Path:         (&file.Finalize{}).Path(),
			Timeout:      0,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &file.Finalize{}
			},
			GetExampleArgs: func() interface{} {
				return &file.Finalize{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Task:    app.ExampleID(),
					ID:      app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleFile
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*file.Finalize)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				epsutil.MustLockProject(tx, args.Host, args.Project)
				epsutil.TaskMustExist(tx, args.Host, args.Project, args.Task)
				f := getOne(tx, args.Host, args.Project, args.Task, args.ID, false)
				app.ReturnIf(f == nil || !f.CreatedBy.Equal(me.Get(tlbx)), http.StatusNotFound, "temp file not found")
				f.CreatedOn = NowMilli()
				srv.Store().MustCopy(cnsts.TempFileBucket, cnsts.FileBucket, store.Key("", args.Host, args.Project, args.Task, args.ID))
				// update new file to finalized status
				_, err := tx.Exec(`UPDATE files SET isFinalized=1, createdOn=? WHERE host=? AND project=? AND task=? AND id=?`, f.CreatedOn, args.Host, args.Project, args.Task, args.ID)
				PanicOn(err)
				// update task file values
				_, err = tx.Exec(`UPDATE tasks SET fileCount=fileCount+1, fileSize=fileSize+? WHERE host=? AND project=? AND id=?`, f.Size, args.Host, args.Project, args.Task)
				PanicOn(err)
				// propogate aggregate sizes upwards
				epsutil.SetAncestralChainAggregateValuesFromParentOfTask(tx, args.Host, args.Project, args.Task)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, &args.Task, f.ID, cnsts.TypeFile, cnsts.ActionCreated, &f.Name, nil)
				tx.Commit()
				return f
			},
		},
		{
			Description:  "Get a presigend get url",
			Path:         (&file.GetPresignedGetUrl{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &file.GetPresignedGetUrl{}
			},
			GetExampleArgs: func() interface{} {
				return &file.GetPresignedGetUrl{
					Host:       app.ExampleID(),
					Project:    app.ExampleID(),
					Task:       app.ExampleID(),
					ID:         app.ExampleID(),
					IsDownload: true,
				}
			},
			GetExampleResponse: func() interface{} {
				return "https://minio.com/get/your/file/here"
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*file.GetPresignedGetUrl)
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				f := getOne(tx, args.Host, args.Project, args.Task, args.ID, true)
				app.ReturnIf(f == nil, http.StatusNotFound, "file not found")
				url := srv.Store().MustPresignedGetUrl(cnsts.FileBucket, store.Key("", args.Host, args.Project, args.Task, args.ID), f.Name, args.IsDownload)
				tx.Commit()
				return url
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
				args.Limit = sql.Limit100(args.Limit)
				res := &file.GetRes{
					Set:  make([]*file.File, 0, args.Limit),
					More: false,
				}
				qry := bytes.NewBufferString(`SELECT task, id, name, createdBy, createdOn, size, mimeType FROM files WHERE isFinalized=1 AND host=? AND project=?`)
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
						qry.WriteString(Strf(` AND createdOn %s= (SELECT f.createdOn FROM files f WHERE f.host=? AND f.project=? AND f.id=?) AND id <> ?`, sql.GtLtSymbol(*args.Asc)))
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
				f := getOne(tx, args.Host, args.Project, args.Task, args.ID, true)
				app.ReturnIf(f == nil, http.StatusNotFound, "file not found")
				app.ReturnIf(role == cnsts.RoleWriter && (!f.CreatedBy.Equal(me) || f.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only delete your own file entries within an hour of creating it")
				// delete file
				_, err := tx.Exec(`DELETE FROM files WHERE host=? AND project=? AND task=? AND id=?`, args.Host, args.Project, args.Task, args.ID)
				PanicOn(err)
				_, err = tx.Exec(`UPDATE tasks SET fileCount=fileCount-1, fileSize=fileSize-? WHERE host=? AND project=? AND id=?`, f.Size, args.Host, args.Project, args.Task)
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
	exampleFile = &file.File{
		Task:      app.ExampleID(),
		ID:        app.ExampleID(),
		CreatedBy: app.ExampleID(),
		CreatedOn: app.ExampleTime(),
		Size:      60,
		MimeType:  "application/pdf",
		Name:      "my_file.pdf",
	}
)

func getOne(tx service.Tx, host, project, task, id ID, isFinalized bool) *file.File {
	row := tx.QueryRow(`SELECT task, id, name, createdBy, createdOn, size, mimeType FROM files WHERE host=? AND project=? AND task=? AND isFinalized=? AND id=?`, host, project, task, isFinalized, id)
	t, err := Scan(row)
	sql.PanicIfIsntNoRows(err)
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
		&f.MimeType)
	if f.ID.IsZero() {
		f = nil
	}
	return f, err
}
