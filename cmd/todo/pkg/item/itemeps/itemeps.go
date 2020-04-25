package itemeps

import (
	"bytes"
	"net/http"
	"time"

	"github.com/0xor1/wtf/cmd/todo/pkg/item"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/field"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/ptr"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
	"github.com/0xor1/wtf/pkg/web/app/common/sql"
	"github.com/0xor1/wtf/pkg/web/app/common/validate"
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
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*item.Create)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				me := tlbx.Me()
				serv := service.Get(tlbx)
				res := &item.Item{
					ID:        tlbx.NewID(),
					CreatedOn: NowMilli(),
					Name:      args.Name,
				}
				tx := serv.Data().Begin()
				defer tx.Rollback()
				_, err := tx.Exec(`INSERT INTO items (user, list, id, createdOn, name, completedOn) VALUES (?, ?, ?, ?, ?, ?)`, me, args.List, res.ID, res.CreatedOn, res.Name, time.Time{})
				PanicOn(err)
				_, err = tx.Exec(`UPDATE lists SET todoItemCount = todoItemCount + 1 WHERE user=? AND id=?`, me, args.List)
				PanicOn(err)
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
					Limit:     ptr.Int(100),
				}
			},
			GetExampleArgs: func() interface{} {
				return &item.Get{
					List:           app.ExampleID(),
					NameStartsWith: ptr.String("My I"),
					CreatedOnMin:   ptr.Time(app.ExampleTime()),
					CreatedOnMax:   ptr.Time(app.ExampleTime()),
					Completed:      ptr.Bool(true),
					CompletedOnMin: ptr.Time(app.ExampleTime()),
					CompletedOnMax: ptr.Time(app.ExampleTime()),
					After:          ptr.ID(app.ExampleID()),
					Sort:           item.SortName,
					Asc:            ptr.Bool(true),
					Limit:          ptr.Int(50),
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
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
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
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*item.Update)
				if args.Name != nil {
					validate.Str("name", args.Name.V, tlbx, nameMinLen, nameMaxLen)
				}
				me := tlbx.Me()
				getSetRes := getSet(tlbx, &item.Get{
					List:  args.List,
					IDs:   IDs{args.ID},
					Limit: ptr.Int(1),
				})
				tlbx.ReturnMsgIf(len(getSetRes.Set) == 0, http.StatusNotFound, "no list with that id")
				item := getSetRes.Set[0]
				changeMade := false
				todoItemCountOp := ""
				completedItemCountOp := ""
				if args.Name != nil && item.Name != args.Name.V {
					item.Name = args.Name.V
					changeMade = true
				}
				if args.Complete != nil &&
					((args.Complete.V && item.CompletedOn == nil) ||
						(!args.Complete.V && item.CompletedOn != nil)) {
					if args.Complete.V {
						item.CompletedOn = ptr.Time(NowMilli())
						todoItemCountOp = "-"
						completedItemCountOp = "+"
					} else {
						item.CompletedOn = nil
						todoItemCountOp = "+"
						completedItemCountOp = "-"
					}
					changeMade = true
				}
				if changeMade {
					serv := service.Get(tlbx)
					tx := serv.Data().Begin()
					defer tx.Rollback()
					sqlRes, err := tx.Exec(`UPDATE items SET name=?, completedOn=? WHERE user=? AND list=? AND id=?`, item.Name, ptr.TimeOr(item.CompletedOn, time.Time{}), me, args.List, item.ID)
					PanicOn(err)
					rowsEffected, err := sqlRes.RowsAffected()
					PanicOn(err)
					if rowsEffected == 1 && todoItemCountOp != "" {
						_, err := tx.Exec(Sprintf(`UPDATE lists SET todoItemCount = todoItemCount %s 1, completedItemCount = completedItemCount %s 1 WHERE user=? AND id=?`, todoItemCountOp, completedItemCountOp), me, args.List)
						PanicOn(err)
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
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*item.Delete)
				idsLen := len(args.IDs)
				if idsLen == 0 {
					return nil
				}
				validate.MaxIDs(tlbx, "ids", args.IDs, 100)
				me := tlbx.Me()
				serv := service.Get(tlbx)
				queryArgs := make([]interface{}, 0, idsLen+2)
				queryArgs = append(queryArgs, me, args.List)
				queryArgs = append(queryArgs, args.IDs.ToIs()...)
				tx := serv.Data().Begin()
				defer tx.Rollback()
				_, err := serv.Data().Exec(`DELETE FROM items WHERE user=? AND list=?`+sql.InCondition(true, "id", idsLen), queryArgs...)
				PanicOn(err)
				_, err = serv.Data().Exec(`UPDATE lists SET todoItemCount = (SELECT COUNT(id) FROM items WHERE user=? AND list=? AND completedOn=?), completedItemCount = (SELECT COUNT(id) FROM items WHERE user=? AND list=? AND completedOn<>?) WHERE user=? AND id=?`, me, args.List, time.Time{}, me, args.List, time.Time{}, me, args.List)
				PanicOn(err)
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

func getSet(tlbx app.Toolbox, args *item.Get) *item.GetRes {
	validate.MaxIDs(tlbx, "ids", args.IDs, 100)
	tlbx.ReturnMsgIf(
		args.CreatedOnMin != nil &&
			args.CreatedOnMax != nil &&
			args.CreatedOnMin.After(*args.CreatedOnMax),
		http.StatusBadRequest, "createdOnMin must be before createdOnMax")
	limit := sql.Limit(*args.Limit, 100)
	me := tlbx.Me()
	serv := service.Get(tlbx)
	res := &item.GetRes{
		Set: make([]*item.Item, 0, limit),
	}
	query := bytes.NewBufferString(`SELECT id, createdOn, name, completedOn FROM items WHERE user=? AND list=?`)
	queryArgs := make([]interface{}, 0, 10)
	queryArgs = append(queryArgs, me, args.List)
	idsLen := len(args.IDs)
	if idsLen > 0 {
		query.WriteString(sql.InCondition(true, `id`, idsLen))
		query.WriteString(sql.OrderByField(`id`, idsLen))
		queryArgs = append(queryArgs, args.IDs.ToIs()...)
		queryArgs = append(queryArgs, args.IDs.ToIs()...)
	} else {
		queryArgs = append(queryArgs, time.Time{})
		if ptr.BoolOr(args.Completed, false) {
			query.WriteString(` AND completedOn <> ?`)
			if args.CompletedOnMin != nil {
				query.WriteString(` AND completedOn >= ?`)
				queryArgs = append(queryArgs, *args.CompletedOnMin)
			}
			if args.CompletedOnMax != nil {
				query.WriteString(` AND completedOn <= ?`)
				queryArgs = append(queryArgs, *args.CompletedOnMax)
			}
		} else {
			query.WriteString(` AND completedOn = ?`)
		}
		if ptr.StringOr(args.NameStartsWith, "") != "" {
			query.WriteString(` AND name LIKE ?`)
			queryArgs = append(queryArgs, Sprintf(`%s%%`, *args.NameStartsWith))
		}
		if args.CreatedOnMin != nil {
			query.WriteString(` AND createdOn >= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMin)
		}
		if args.CreatedOnMax != nil {
			query.WriteString(` AND createdOn <= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMax)
		}
		if args.After != nil {
			query.WriteString(Sprintf(` AND %s %s= (SELECT %s FROM items WHERE user=? AND list=? AND id=?) AND id <> ?`, args.Sort, sql.GtLtSymbol(*args.Asc), args.Sort))
			queryArgs = append(queryArgs, me, args.List, *args.After, *args.After)
			if args.Sort != item.SortCreatedOn {
				query.WriteString(Sprintf(` AND createdOn %s (SELECT createdOn FROM items WHERE user=? AND list=? AND id=?)`, sql.GtLtSymbol(*args.Asc)))
				queryArgs = append(queryArgs, me, args.List, *args.After)

			}
		}
		createdOnSecondarySort := ""
		if args.Sort != item.SortCreatedOn {
			createdOnSecondarySort = ", createdOn"
		}
		query.WriteString(Sprintf(` ORDER BY %s%s %s, id LIMIT %d`, args.Sort, createdOnSecondarySort, sql.Asc(*args.Asc), limit))
	}
	serv.Data().Query(func(rows isql.Rows) {
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == limit {
				res.More = true
				break
			}
			i := &item.Item{}
			completedOn := time.Time{}
			PanicOn(rows.Scan(&i.ID, &i.CreatedOn, &i.Name, &completedOn))
			if !completedOn.IsZero() {
				i.CompletedOn = &completedOn
			}
			res.Set = append(res.Set, i)
		}
	}, query.String(), queryArgs...)
	return res
}
