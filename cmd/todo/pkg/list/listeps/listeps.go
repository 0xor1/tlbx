package listeps

//go:generate go install github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=listeps.sql -skipLineComments

import (
	"net/http"

	"github.com/0xor1/sqlx"
	"github.com/0xor1/tlbx/cmd/todo/pkg/list"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/sqlh"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/filter"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new list",
			Path:         (&list.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &list.Create{
					Name: "My List",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleList
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*list.Create)
				validate.Str("name", args.Name, nameMinLen, nameMaxLen)
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				res := &list.List{
					ID:                 tlbx.NewID(),
					CreatedOn:          NowMilli(),
					Name:               args.Name,
					TodoItemCount:      0,
					CompletedItemCount: 0,
				}
				srv.Data().MustExec(qryListInsert(),
					me, res.ID, res.CreatedOn, res.Name, res.TodoItemCount, res.CompletedItemCount)
				return res
			},
		},
		{
			Description:  "Get a list set",
			Path:         (&list.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Get{
					Base: FilterDefs(),
				}
			},
			GetExampleArgs: func() interface{} {
				return &list.Get{
					NamePrefix:            ptr.String("My L"),
					CreatedOnMin:          ptr.Time(app.ExampleTime()),
					CreatedOnMax:          ptr.Time(app.ExampleTime()),
					TodoItemCountMin:      ptr.Int(2),
					TodoItemCountMax:      ptr.Int(5),
					CompletedItemCountMin: ptr.Int(3),
					CompletedItemCountMax: ptr.Int(4),
					Base: filter.Base{
						After: ptr.ID(app.ExampleID()),
						Sort:  list.SortName,
						Asc:   ptr.Bool(true),
						Limit: 50,
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return &list.GetRes{
					Set: []*list.List{
						exampleList,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return getSet(tlbx, a.(*list.Get))
			},
		},
		{
			Description:  "Update a list",
			Path:         (&list.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &list.Update{
					ID:   app.ExampleID(),
					Name: field.String{V: "New List Name"},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleList
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*list.Update)
				validate.Str("name", args.Name.V, nameMinLen, nameMaxLen)
				getSetRes := getSet(tlbx, &list.Get{
					Base: filter.Base{
						IDs: IDs{args.ID},
					},
				})
				app.ReturnIf(len(getSetRes.Set) == 0, http.StatusNotFound, "no list with that id")
				list := getSetRes.Set[0]
				list.Name = args.Name.V
				srv := service.Get(tlbx)
				srv.Data().MustExec(qryListUpdate(), list.Name, me.AuthedGet(tlbx), list.ID)
				return list
			},
		},
		{
			Description:  "Delete lists",
			Path:         (&list.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &list.Delete{
					IDs: []ID{app.ExampleID()},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*list.Delete)
				idsLen := len(args.IDs)
				if idsLen == 0 {
					return nil
				}
				validate.MaxIDs("ids", args.IDs, 100)
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				queryArgs := make([]interface{}, 0, idsLen+1)
				queryArgs = append(queryArgs, me)
				queryArgs = append(queryArgs, args.IDs.ToIs()...)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				// items deleted on foreign key cascade
				tx.MustExec(qryListsDelete(idsLen), queryArgs...)
				tx.Commit()
				return nil
			},
		},
	}
	nameMinLen  = 1
	nameMaxLen  = 250
	exampleList = &list.List{
		ID:                 app.ExampleID(),
		Name:               "My List",
		CreatedOn:          app.ExampleTime(),
		TodoItemCount:      3,
		CompletedItemCount: 4,
	}
)

func OnDelete(tlbx app.Tlbx, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().BeginWrite()
	defer tx.Rollback()
	// items deleted on foreign key cascade
	tx.MustExec(qryOnDelete(), me)
	tx.Commit()
}

func getSet(tlbx app.Tlbx, args *list.Get) *list.GetRes {
	app.BadReqIf(
		args.CreatedOnMin != nil &&
			args.CreatedOnMax != nil &&
			args.CreatedOnMin.After(*args.CreatedOnMax),
		"createdOnMin must be before createdOnMax")
	app.BadReqIf(
		args.TodoItemCountMin != nil &&
			args.TodoItemCountMax != nil &&
			*args.TodoItemCountMin > *args.TodoItemCountMax,
		"todoItemCountMin must not be greater than todoItemCountMax")
	app.BadReqIf(
		args.CompletedItemCountMin != nil &&
			args.CompletedItemCountMax != nil &&
			*args.CompletedItemCountMin > *args.CompletedItemCountMax,
		"completedItemCountMin must not be greater than completedItemCountMax")
	args.Base.Limit = sqlh.Limit100(args.Base.Limit)
	me := me.AuthedGet(tlbx)
	srv := service.Get(tlbx)
	res := &list.GetRes{
		Set: make([]*list.List, 0, args.Base.Limit),
	}
	sqlArgs := sqlh.NewArgs(10)
	srv.Data().MustQuery(func(rows *sqlx.Rows) {
		iLimit := int(args.Base.Limit)
		for rows.Next() {
			if len(args.Base.IDs) == 0 && len(res.Set)+1 == iLimit {
				res.More = true
				break
			}
			l := &list.List{}
			PanicOn(rows.Scan(&l.ID, &l.CreatedOn, &l.Name, &l.TodoItemCount, &l.CompletedItemCount))
			res.Set = append(res.Set, l)
		}
	}, qryListsGet(sqlArgs, me, args), sqlArgs.Is()...)
	return res
}

func FilterDefs() filter.Base {
	return filter.DefsAsc100(
		list.SortCreatedOn,
		list.SortName,
		list.SortTodoItemCount,
		list.SortCompletedItemCount,
	)
}
