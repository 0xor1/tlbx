package itemeps

//go:generate go install github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=itemeps.sql -skipLineComments

import (
	"net/http"
	"time"

	"github.com/0xor1/tlbx/cmd/todo/pkg/item"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/sqlh"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new item",
			Path:         (&item.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &item.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &item.Create{
					List: app.ExampleID(),
					Name: "My Item",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleItem
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*item.Create)
				validate.Str("name", args.Name, nameMinLen, nameMaxLen)
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				res := &item.Item{
					ID:        tlbx.NewID(),
					CreatedOn: NowMilli(),
					Name:      args.Name,
				}
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				tx.MustExec(qryItemInsert(), me, args.List, res.ID, res.CreatedOn, res.Name, time.Time{})
				tx.MustExec(qryIncrementListItemCount(), me, args.List)
				tx.Commit()
				return res
			},
		},
		{
			Description:  "Get an item set",
			Path:         (&item.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &item.Get{
					Completed: ptr.Bool(false),
					Sort:      item.SortCreatedOn,
					Asc:       ptr.Bool(true),
					Limit:     100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &item.Get{
					List:           app.ExampleID(),
					NamePrefix:     ptr.String("My I"),
					CreatedOnMin:   ptr.Time(app.ExampleTime()),
					CreatedOnMax:   ptr.Time(app.ExampleTime()),
					Completed:      ptr.Bool(true),
					CompletedOnMin: ptr.Time(app.ExampleTime()),
					CompletedOnMax: ptr.Time(app.ExampleTime()),
					After:          ptr.ID(app.ExampleID()),
					Sort:           item.SortName,
					Asc:            ptr.Bool(true),
					Limit:          50,
				}
			},
			GetExampleResponse: func() interface{} {
				return &item.GetRes{
					Set: []*item.Item{
						exampleItem,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return getSet(tlbx, a.(*item.Get))
			},
		},
		{
			Description:  "Update an item",
			Path:         (&item.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &item.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &item.Update{
					List:     app.ExampleID(),
					ID:       app.ExampleID(),
					Name:     &field.String{V: "New Item Name"},
					Complete: &field.Bool{V: true},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleItem
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*item.Update)
				if args.Name != nil {
					validate.Str("name", args.Name.V, nameMinLen, nameMaxLen)
				}
				me := me.AuthedGet(tlbx)
				getSetRes := getSet(tlbx, &item.Get{
					List: args.List,
					IDs:  IDs{args.ID},
				})
				app.ReturnIf(len(getSetRes.Set) == 0, http.StatusNotFound, "no list with that id")
				item := getSetRes.Set[0]
				changeMade := false
				countsChanged := false
				if args.Name != nil && item.Name != args.Name.V {
					item.Name = args.Name.V
					changeMade = true
				}
				if args.Complete != nil &&
					((args.Complete.V && item.CompletedOn == nil) ||
						(!args.Complete.V && item.CompletedOn != nil)) {
					changeMade = true
					countsChanged = true
					if args.Complete.V {
						item.CompletedOn = ptr.Time(NowMilli())
					} else {
						item.CompletedOn = nil
					}
				}
				if changeMade {
					srv := service.Get(tlbx)
					tx := srv.Data().BeginWrite()
					defer tx.Rollback()
					sqlRes := tx.MustExec(qryItemUpdate(), item.Name, ptr.TimeOr(item.CompletedOn, time.Time{}), me, args.List, item.ID)
					rowsEffected, err := sqlRes.RowsAffected()
					PanicOn(err)
					if rowsEffected == 1 && countsChanged {
						tx.MustExec(qryListCountsToggle(args.Complete.V), me, args.List)
					}
					tx.Commit()
				}
				return item
			},
		},
		{
			Description:  "Delete items",
			Path:         (&item.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &item.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &item.Delete{
					List: app.ExampleID(),
					IDs:  []ID{app.ExampleID()},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*item.Delete)
				idsLen := len(args.IDs)
				if idsLen == 0 {
					return nil
				}
				validate.MaxIDs("ids", args.IDs, 100)
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				queryArgs := make([]interface{}, 0, idsLen+2)
				queryArgs = append(queryArgs, me, args.List)
				queryArgs = append(queryArgs, args.IDs.ToIs()...)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				tx.MustExec(qryItemsDelete(len(args.IDs)), queryArgs...)
				tx.MustExec(qryListRecalculateCounts(), me, args.List, time.Time{}, me, args.List, time.Time{}, me, args.List)
				tx.Commit()
				return nil
			},
		},
	}
	nameMinLen  = 1
	nameMaxLen  = 250
	exampleItem = &item.Item{
		ID:        app.ExampleID(),
		Name:      "My Item",
		CreatedOn: app.ExampleTime(),
	}
)

func getSet(tlbx app.Tlbx, args *item.Get) *item.GetRes {
	validate.MaxIDs("ids", args.IDs, 100)
	app.BadReqIf(
		args.CreatedOnMin != nil &&
			args.CreatedOnMax != nil &&
			args.CreatedOnMin.After(*args.CreatedOnMax),
		"createdOnMin must be before createdOnMax")
	args.Limit = sqlh.Limit100(args.Limit)
	me := me.AuthedGet(tlbx)
	srv := service.Get(tlbx)
	res := &item.GetRes{
		Set: make([]*item.Item, 0, args.Limit),
	}
	sqlArgs := &sqlh.Args{}
	srv.Data().MustGetN(&res.Set, qryItemsGet(sqlArgs, me, args), sqlArgs.Is()...)
	if len(args.IDs) == 0 && len(res.Set) == int(args.Limit) {
		res.Set = res.Set[:len(res.Set)-1]
		res.More = true
	}
	for _, i := range res.Set {
		if i.CompletedOn.IsZero() {
			i.CompletedOn = nil
		}
	}
	return res
}
