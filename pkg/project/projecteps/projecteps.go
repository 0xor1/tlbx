package projecteps

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/0xor1/trees/pkg/epsutil"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new project",
			Path:         (&project.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Create{
					CurrencyCode: "USD",
					HoursPerDay:  nil,
					DaysPerWeek:  nil,
					IsPublic:     false,
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Create{
					CurrencyCode: "USD",
					HoursPerDay:  ptr.Uint8(8),
					DaysPerWeek:  ptr.Uint8(5),
					StartOn:      ptr.Time(app.ExampleTime()),
					EndOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
					IsPublic:     false,
					Name:         "My New Project",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleProject
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.Create)
				me := me.AuthedGet(tlbx)
				args.Name = StrTrimWS(args.Name)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				if args.CurrencyCode == "" {
					args.CurrencyCode = "USD"
				}
				validateCurrencyCode(tlbx, args.CurrencyCode)
				app.BadReqIf(args.HoursPerDay != nil && (*args.HoursPerDay < 1 || *args.HoursPerDay > 24), "invalid hoursPerDay must be > 0 and <= 24")
				app.BadReqIf(args.DaysPerWeek != nil && (*args.DaysPerWeek < 1 || *args.DaysPerWeek > 7), "invalid daysPerWeek must be > 0 and <= 7")
				app.BadReqIf((args.HoursPerDay == nil && args.DaysPerWeek != nil) ||
					(args.HoursPerDay != nil && args.DaysPerWeek == nil), "invalid hoursPerDay and daysPerWeek must both be set or not set")
				app.BadReqIf(args.StartOn != nil && args.EndOn != nil && !args.StartOn.Before(*args.EndOn), "invalid startOn must be before endOn")
				p := &project.Project{
					Task: task.Task{
						ID:         tlbx.NewID(),
						Name:       args.Name,
						User:       ptr.ID(me),
						CreatedBy:  me,
						CreatedOn:  tlbx.Start(),
						IsParallel: false,
					},
					Base: project.Base{
						CurrencyCode: args.CurrencyCode,
						HoursPerDay:  args.HoursPerDay,
						DaysPerWeek:  args.DaysPerWeek,
						StartOn:      args.StartOn,
						EndOn:        args.EndOn,
						IsPublic:     args.IsPublic,
					},
					Host:       me,
					IsArchived: false,
				}
				srv := service.Get(tlbx)

				u := &user.User{}
				row := srv.User().QueryRow(`SELECT id, handle, alias, hasAvatar FROM users WHERE id=?`, me)
				PanicOn(row.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar))

				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				_, err := tx.Exec(`INSERT INTO projectLocks (host, id) VALUES (?, ?)`, p.Host, p.ID)
				PanicOn(err)
				_, err = tx.Exec(`INSERT INTO users (host, project, id, handle, alias, hasAvatar, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, p.Host, p.ID, me, u.Handle, u.Alias, u.HasAvatar, true, cnsts.RoleAdmin)
				PanicOn(err)
				_, err = tx.Exec(`INSERT INTO projects (host, id, isArchived, name, createdOn, currencyCode, hoursPerDay, daysPerWeek, startOn, endOn, isPublic, fileLimit) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.Host, p.ID, p.IsArchived, p.Name, p.CreatedOn, p.CurrencyCode, p.HoursPerDay, p.DaysPerWeek, p.StartOn, p.EndOn, p.IsPublic, p.FileLimit)
				PanicOn(err)
				_, err = tx.Exec(`INSERT INTO tasks (host, project, id, parent, firstChild, nextSib, user, name, description, isParallel, createdBy, createdOn, timeEst, timeInc, timeSubMin, timeSubEst, timeSubInc, costEst, costInc, costSubEst, costSubInc, fileN, fileSize, fileSubN, fileSubSize, childN, descN) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.Host, p.ID, p.ID, p.Parent, p.FirstChild, p.NextSib, p.User, p.Name, p.Description, p.IsParallel, p.CreatedBy, p.CreatedOn, p.TimeEst, p.TimeInc, p.TimeSubMin, p.TimeSubEst, p.TimeSubInc, p.CostEst, p.CostInc, p.CostSubEst, p.CostSubInc, p.FileN, p.FileSize, p.FileSubN, p.FileSubSize, p.ChildN, p.DescN)
				PanicOn(err)
				epsutil.LogActivity(tlbx, tx, me, p.ID, p.ID, p.ID, cnsts.TypeTask, cnsts.ActionCreated, ptr.String(p.Name), nil, nil, nil)
				tx.Commit()
				return p
			},
		},
		{
			Description:  "Get a project set",
			Path:         (&project.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Get{
					IsArchived: false,
					Sort:       cnsts.SortCreatedOn,
					Asc:        ptr.Bool(true),
					Limit:      100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Get{
					IsArchived:   false,
					Others:       true,
					NamePrefix:   ptr.String("My Proj"),
					CreatedOnMin: ptr.Time(app.ExampleTime()),
					CreatedOnMax: ptr.Time(app.ExampleTime()),
					After:        ptr.ID(app.ExampleID()),
					Sort:         cnsts.SortName,
					Asc:          ptr.Bool(true),
					Limit:        50,
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.GetRes{
					Set: []*project.Project{
						exampleProject,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return getSet(tlbx, a.(*project.Get))
			},
		},
		{
			Description:  "Get latest public projects",
			Path:         (&project.GetLatestPublic{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return &project.GetLatestPublicRes{
					Set: []*project.Project{
						exampleProject,
					},
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				res := &project.GetLatestPublicRes{}
				tmpRes := getLatestPublicProjects(tlbx)
				if tmpRes != nil {
					res.Set = tmpRes.Set
				}
				return res
			},
		},
		{
			Description:  "Update a project",
			Path:         (&project.Updates{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Updates{}
			},
			GetExampleArgs: func() interface{} {
				return &project.Updates{
					{
						ID:           app.ExampleID(),
						Name:         &field.String{V: "Renamed Project"},
						CurrencyCode: &field.String{V: "EUR"},
						HoursPerDay:  &field.UInt8Ptr{V: ptr.Uint8(6)},
						DaysPerWeek:  &field.UInt8Ptr{V: ptr.Uint8(4)},
						StartOn:      &field.TimePtr{V: ptr.Time(app.ExampleTime())},
						EndOn:        &field.TimePtr{V: ptr.Time(app.ExampleTime().Add(24 * time.Hour))},
						IsArchived:   &field.Bool{V: false},
						IsPublic:     &field.Bool{V: true},
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return []*project.Project{
					exampleProject,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := *(a.(*project.Updates))
				if len(args) == 0 {
					return nil
				}
				app.BadReqIf(len(args) > 100, "can not update more than 100 projects at a time")
				me := me.AuthedGet(tlbx)
				ids := make(IDs, 0, len(args))
				namesSet := make([]bool, len(args))
				dupes := map[string]bool{}
				for i := 0; i < len(args); i++ {
					u := args[i]
					// if there are no changes to be made, remove this entry
					if u.Name == nil &&
						u.CurrencyCode == nil &&
						u.HoursPerDay == nil &&
						u.DaysPerWeek == nil &&
						u.StartOn == nil &&
						u.EndOn == nil &&
						u.IsArchived == nil &&
						u.IsPublic == nil {
						copy(args[i:], args[i+1:])
						args[len(args)-1] = nil
						args = args[:len(args)-1]
						i--
					} else {
						idStr := u.ID.String()
						app.BadReqIf(dupes[idStr], "duplicate entry detected")
						dupes[idStr] = true
						ids = append(ids, u.ID)
					}
				}
				ps := getSet(tlbx, &project.Get{Host: me, IDs: ids}).Set
				for i, p := range ps {
					a := args[i]
					if a.CurrencyCode != nil {
						validateCurrencyCode(tlbx, a.CurrencyCode.V)
						p.CurrencyCode = a.CurrencyCode.V
					}
					// validate name
					if a.Name != nil {
						a.Name.V = StrTrimWS(a.Name.V)
						validate.Str("name", a.Name.V, tlbx, nameMinLen, nameMaxLen)
						p.Name = a.Name.V
						namesSet[i] = true
					}
					// validate startOn and endOn
					switch {
					case a.StartOn != nil && a.EndOn != nil:
						app.BadReqIf(a.StartOn.V != nil && a.EndOn.V != nil && !a.StartOn.V.Before(*a.EndOn.V), "invalid startOn must be before endOn")
					case a.StartOn != nil && p.EndOn != nil:
						app.BadReqIf(a.StartOn.V != nil && p.EndOn != nil && !a.StartOn.V.Before(*p.EndOn), "invalid startOn must be before endOn")
					case a.EndOn != nil && p.StartOn != nil:
						app.BadReqIf(p.StartOn != nil && a.EndOn.V != nil && !p.StartOn.Before(*a.EndOn.V), "invalid startOn must be before endOn")
					}
					if a.StartOn != nil {
						p.StartOn = a.StartOn.V
					}
					if a.EndOn != nil {
						p.EndOn = a.EndOn.V
					}
					if a.HoursPerDay != nil {
						app.BadReqIf(a.HoursPerDay.V != nil && (*a.HoursPerDay.V < 1 || *a.HoursPerDay.V > 24), "invalid hoursPerDay must be > 0 and <= 24")
						p.HoursPerDay = a.HoursPerDay.V
					}
					if a.DaysPerWeek != nil {
						app.BadReqIf(a.DaysPerWeek.V != nil && (*a.DaysPerWeek.V < 1 || *a.DaysPerWeek.V > 7), "invalid daysPerWeek must be > 0 and <= 7")
						p.DaysPerWeek = a.DaysPerWeek.V
					}
					app.BadReqIf(
						(p.HoursPerDay == nil && p.DaysPerWeek != nil) ||
							(p.HoursPerDay != nil && p.DaysPerWeek == nil),
						"invalid hoursPerDay And daysPerWeek must both be either set or not set")
					if a.IsArchived != nil {
						p.IsArchived = a.IsArchived.V
					}
					if a.IsPublic != nil {
						p.IsPublic = a.IsPublic.V
					}
				}
				srv := service.Get(tlbx)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				for i, p := range ps {
					_, err := tx.Exec(`UPDATE projects SET name=?, currencyCode=?, hoursPerDay=?, daysPerWeek=?, startOn=?, endOn=?, isArchived=?, isPublic=? WHERE host=? AND id=?`, p.Name, p.CurrencyCode, p.HoursPerDay, p.DaysPerWeek, p.StartOn, p.EndOn, p.IsArchived, p.IsPublic, me, p.ID)
					PanicOn(err)
					if namesSet[i] {
						_, err = tx.Exec(`UPDATE tasks SET name=? WHERE host=? AND project=? AND id=?`, p.Name, me, p.ID, p.ID)
						PanicOn(err)
						epsutil.ActivityItemRename(tx, me, p.ID, p.ID, p.Name, true)
					}
					epsutil.LogActivity(tlbx, tx, me, p.ID, p.ID, p.ID, cnsts.TypeTask, cnsts.ActionUpdated, ptr.String(p.Name), args[i], nil, nil)
				}
				tx.Commit()
				return ps
			},
		},
		{
			Description:  "delete projects",
			Path:         (&project.Delete{}).Path(),
			Timeout:      0,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &project.Delete{
					app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := *(a.(*project.Delete))
				if len(args) == 0 {
					return nil
				}
				app.BadReqIf(len(args) > 100, "can not delete more than 100 projects at a time")
				me := me.AuthedGet(tlbx)
				queryArgs := append([]interface{}{me}, IDs(args).ToIs()...)
				inID := sql.InCondition(true, "id", len(args))
				inProject := sql.InCondition(true, "project", len(args))
				srv := service.Get(tlbx)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				_, err := tx.Exec(Strf(`DELETE FROM projectLocks WHERE host=? %s`, inID), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM users WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM activities WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM projects WHERE host=? %s`, inID), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM tasks WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM vitems WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM files WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM comments WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				for _, p := range args {
					srv.Store().MustDeletePrefix(cnsts.FileBucket, epsutil.StorePrefix(me, p))
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "add project users",
			Path:         (&project.AddUsers{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.AddUsers{}
			},
			GetExampleArgs: func() interface{} {
				return &project.AddUsers{
					Project: app.ExampleID(),
					Users: []*project.SendUser{
						{
							ID:   app.ExampleID(),
							Role: cnsts.RoleAdmin,
						},
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.AddUsers)
				lenUsers := len(args.Users)
				if lenUsers == 0 {
					return nil
				}
				app.BadReqIf(lenUsers > 100, "can not add more than 100 users to a project at a time")
				srv := service.Get(tlbx)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleAdmin)

				// need two sets for id IN (?, ...) and ORDER BY FIELD (id, ?, ...)
				ids := make([]interface{}, 0, 2*lenUsers)
				for i := 0; i < 2; i++ {
					for _, u := range args.Users {
						ids = append(ids, u.ID)
					}
				}
				// get userTx and lock all user rows, to ensure they are not changed whilst inserting into data db
				userTx := srv.User().BeginWrite()
				defer userTx.Rollback()
				users := make([]*project.User, 0, lenUsers)
				PanicOn(userTx.Query(func(rows isql.Rows) {
					for rows.Next() {
						u := &project.User{}
						PanicOn(rows.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar))
						u.IsActive = true
						users = append(users, u)
					}
				}, Strf(`SELECT id, handle, alias, hasAvatar FROM users WHERE 1=1 %s %s FOR UPDATE`, sql.InCondition(true, `id`, lenUsers), sql.OrderByField(`id`, lenUsers)), ids...))

				app.BadReqIf(len(users) != lenUsers, "users specified: %d, users found: %d", lenUsers, len(users))

				for i, u := range users {
					app.BadReqIf(u.ID.Equal(args.Host), "can not add host to project")
					u.Role = args.Users[i].Role
					_, err := tx.Exec(`INSERT INTO users (host, project, id, handle, alias, hasAvatar, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, u.ID, u.Handle, u.Alias, u.HasAvatar, u.IsActive, u.Role)
					PanicOn(err)
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Project, u.ID, cnsts.TypeUser, cnsts.ActionCreated, nil, u.Role, nil, nil)
				}
				tx.Commit()
				userTx.Commit()
				return nil
			},
		},
		{
			Description:  "get my project user",
			Path:         (&project.GetMe{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.GetMe{}
			},
			GetExampleArgs: func() interface{} {
				return &project.GetMe{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleUser
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				if !me.AuthedExists(tlbx) {
					return nil
				}
				args := a.(*project.GetMe)
				users := getUsers(tlbx, &project.GetUsers{
					Host:    args.Host,
					Project: args.Project,
					IDs:     IDs{me.AuthedGet(tlbx)},
				})
				if len(users.Set) == 0 {
					return nil
				}
				return users.Set[0]
			},
		},
		{
			Description:  "get project users",
			Path:         (&project.GetUsers{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.GetUsers{
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				r := cnsts.RoleReader
				return &project.GetUsers{
					Host:         app.ExampleID(),
					Project:      app.ExampleID(),
					IDs:          IDs{app.ExampleID()},
					Role:         &r,
					HandlePrefix: ptr.String("my_frien"),
					After:        ptr.ID(app.ExampleID()),
					Limit:        100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.GetUsersRes{
					Set: []*project.User{
						exampleUser,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return getUsers(tlbx, a.(*project.GetUsers))
			},
		},
		{
			Description:  "set project user roles",
			Path:         (&project.SetUserRoles{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.SetUserRoles{}
			},
			GetExampleArgs: func() interface{} {
				return &project.SetUserRoles{
					Project: app.ExampleID(),
					Users: []*project.SendUser{
						{
							ID:   app.ExampleID(),
							Role: cnsts.RoleAdmin,
						},
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.SetUserRoles)
				lenUsers := len(args.Users)
				if lenUsers == 0 {
					return nil
				}
				app.BadReqIf(lenUsers > 100, "can not set more than 100 user roles in a project at a time")
				srv := service.Get(tlbx)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleAdmin)
				for _, u := range args.Users {
					app.ReturnIf(u.ID.Equal(args.Host), http.StatusForbidden, "can not set hosts role")
					res, err := tx.Exec(`UPDATE users SET role=? WHERE host=? AND project=? AND id=?`, u.Role, args.Host, args.Project, u.ID)
					PanicOn(err)
					count, err := res.RowsAffected()
					PanicOn(err)
					app.ReturnIf(count != 1, http.StatusNotFound, "user: %s not found", u.ID)
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Project, u.ID, cnsts.TypeUser, cnsts.ActionUpdated, nil, u.Role, nil, nil)
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "remove project users",
			Path:         (&project.RemoveUsers{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.RemoveUsers{}
			},
			GetExampleArgs: func() interface{} {
				return &project.RemoveUsers{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Users: IDs{
						app.ExampleID(),
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.RemoveUsers)
				if len(args.Users) == 0 {
					return nil
				}
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				if !(len(args.Users) == 1 &&
					me.Equal(args.Users[0]) &&
					!me.Equal(args.Host)) {
					// here the user is requesting to remove themselves
					// from someone elses project which they can always do
					epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleAdmin)
				}
				queryArgs := make([]interface{}, 0, len(args.Users)+2)
				queryArgs = append(queryArgs, args.Host, args.Project)
				for _, u := range args.Users {
					app.BadReqIf(u.Equal(args.Host), "can not remove host from project")
					queryArgs = append(queryArgs, u)
				}
				_, err := tx.Exec(Strf(`UPDATE users SET isActive=0 WHERE host=? AND project=? %s`, sql.InCondition(true, `id`, len(args.Users))), queryArgs...)
				PanicOn(err)
				_, err = srv.User().Exec(Strf(`DELETE FROM fcmTokens WHERE 1=1 %s`, sql.InCondition(true, `user`, len(args.Users))), args.Users.ToIs()...)
				PanicOn(err)
				for _, u := range args.Users {
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.Project, u, cnsts.TypeUser, cnsts.ActionDeleted, nil, nil, nil, nil)
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "get project activities",
			Path:         (&project.GetActivities{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.GetActivities{
					ExcludeDeletedItems: false,
					Limit:               100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.GetActivities{
					Host:                app.ExampleID(),
					Project:             app.ExampleID(),
					ExcludeDeletedItems: true,
					Task:                ptr.ID(app.ExampleID()),
					Item:                ptr.ID(app.ExampleID()),
					User:                ptr.ID(app.ExampleID()),
					OccuredAfter:        ptr.Time(app.ExampleTime()),
					OccuredBefore:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
					Limit:               100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.GetActivitiesRes{
					Set: []*project.Activity{
						{
							Task:        ptr.ID(app.ExampleID()),
							OccurredOn:  app.ExampleTime(),
							User:        app.ExampleID(),
							Item:        app.ExampleID(),
							ItemType:    cnsts.TypeTask,
							TaskDeleted: true,
							ItemDeleted: false,
							Action:      cnsts.ActionUpdated,
							ItemName:    ptr.String("my task"),
							ExtraInfo:   json.MustFromString(`{"isParallel":true}`),
						},
					},
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.GetActivities)
				args.Limit = sql.Limit100(args.Limit)
				app.BadReqIf(args.OccuredAfter != nil && args.OccuredBefore != nil, "only one of occurredBefore or occurredAfter may be used")
				tx := service.Get(tlbx).Data().BeginRead()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
				query := bytes.NewBufferString(`SELECT occurredOn, user, task, item, itemType, taskDeleted, itemDeleted, action, taskName, itemName, extraInfo FROM activities WHERE host=? AND project=?`)
				queryArgs := make([]interface{}, 0, 7)
				queryArgs = append(queryArgs, args.Host, args.Project)
				if args.ExcludeDeletedItems {
					query.WriteString(` AND itemDeleted=0`)
				}
				if args.Item != nil {
					query.WriteString(` AND item=?`)
					queryArgs = append(queryArgs, *args.Item)
				}
				if args.User != nil {
					query.WriteString(` AND user=?`)
					queryArgs = append(queryArgs, *args.User)
				}
				asc := false
				if args.OccuredAfter != nil {
					asc = true
					query.WriteString(` AND occurredOn>?`)
					queryArgs = append(queryArgs, *args.OccuredAfter)
				}
				if args.OccuredBefore != nil {
					query.WriteString(` AND occurredOn<?`)
					queryArgs = append(queryArgs, *args.OccuredBefore)
				}
				query.WriteString(sql.OrderLimit100(`occurredOn`, asc, args.Limit))
				res := &project.GetActivitiesRes{
					Set: make([]*project.Activity, 0, args.Limit),
				}
				PanicOn(tx.Query(func(rows isql.Rows) {
					iLimit := int(args.Limit)
					for rows.Next() {
						if len(res.Set)+1 == iLimit {
							res.More = true
							break
						}
						pa := &project.Activity{}
						var extraInfo *string
						PanicOn(rows.Scan(&pa.OccurredOn, &pa.User, &pa.Task, &pa.Item, &pa.ItemType, &pa.TaskDeleted, &pa.ItemDeleted, &pa.Action, &pa.TaskName, &pa.ItemName, &extraInfo))
						if extraInfo != nil {
							pa.ExtraInfo = json.MustFromString(*extraInfo)
						}
						res.Set = append(res.Set, pa)
					}
				}, query.String(), queryArgs...))
				tx.Commit()
				return res
			},
		},
	}

	nameMinLen     = 1
	nameMaxLen     = 250
	exampleProject = &project.Project{
		Task: task.Task{
			ID:        app.ExampleID(),
			Name:      "My Project",
			CreatedOn: app.ExampleTime(),
		},
		Base: project.Base{
			CurrencyCode: "USD",
			HoursPerDay:  ptr.Uint8(8),
			DaysPerWeek:  ptr.Uint8(5),
			StartOn:      ptr.Time(app.ExampleTime()),
			EndOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
			IsPublic:     false,
		},
		Host:       app.ExampleID(),
		IsArchived: false,
	}
	exampleUser = &project.User{
		User: user.User{
			ID:        app.ExampleID(),
			Handle:    ptr.String("joe_bloggs"),
			Alias:     ptr.String("joe soap"),
			HasAvatar: ptr.Bool(true),
		},
		Role:     cnsts.RoleReader,
		IsActive: true,
		TimeEst:  13,
		TimeInc:  13,
		CostEst:  13,
		CostInc:  13,
		FileN:    13,
		FileSize: 13,
		TaskN:    13,
	}
)

func OnDelete(tlbx app.Tlbx, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().BeginWrite()
	defer tx.Rollback()
	_, err := tx.Exec(`DELETE FROM projectLocks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM users WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM activities WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projects WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM tasks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM vitems WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM files WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM comments WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`UPDATE users set isActive=0 WHERE id=?`, me)
	PanicOn(err)
	srv.Store().MustDeletePrefix(cnsts.FileBucket, epsutil.StorePrefix(me))
	tx.Commit()
}

func OnSetSocials(tlbx app.Tlbx, user *user.User) {
	srv := service.Get(tlbx)
	tx := srv.Data().BeginWrite()
	defer tx.Rollback()
	_, err := tx.Exec(`UPDATE users SET handle=?, alias=?, hasAvatar=? WHERE id=?`, user.Handle, user.Alias, user.HasAvatar, user.ID)
	PanicOn(err)
	tx.Commit()
}

func ValidateFCMTopic(tlbx app.Tlbx, topic IDs) (sqlh.Tx, error) {
	app.BadReqIf(len(topic) != 2, "fcm topic must be 2 ids, host then project")
	tx := service.Get(tlbx).Data().BeginRead()
	epsutil.IMustHaveAccess(tlbx, tx, topic[0], topic[1], cnsts.RoleReader)
	return tx, nil
}

func GetOne(tlbx app.Tlbx, host, id ID) *project.Project {
	res := getSet(tlbx, &project.Get{Host: host, IDs: IDs{id}})
	if res != nil && len(res.Set) == 1 {
		return res.Set[0]
	}
	return nil
}

func getSet(tlbx app.Tlbx, args *project.Get) *project.GetRes {
	validate.MaxIDs(tlbx, "ids", args.IDs, 100)
	app.BadReqIf(
		args.CreatedOnMin != nil &&
			args.CreatedOnMax != nil &&
			args.CreatedOnMin.After(*args.CreatedOnMax),
		"createdOnMin must be before createdOnMax")
	app.BadReqIf(
		args.StartOnMin != nil &&
			args.StartOnMax != nil &&
			args.StartOnMin.After(*args.StartOnMax),
		"startOnMin must be before startOnMax")
	app.BadReqIf(
		args.EndOnMin != nil &&
			args.EndOnMax != nil &&
			args.EndOnMin.After(*args.EndOnMax),
		"endOnMin must be before endOnMax")
	app.BadReqIf(
		args.StartOnMin != nil &&
			args.EndOnMin != nil &&
			args.StartOnMin.After(*args.EndOnMax),
		"startOnMin must be before endOnMin")
	app.BadReqIf(
		args.StartOnMax != nil &&
			args.EndOnMax != nil &&
			args.StartOnMax.After(*args.EndOnMax),
		"startOnMax must be before endOnMax")
	args.Limit = sql.Limit100(args.Limit)
	srv := service.Get(tlbx)
	res := &project.GetRes{
		Set: make([]*project.Project, 0, args.Limit),
	}
	query := bytes.NewBufferString(projects_select_columns)
	queryArgs := make([]interface{}, 0, 14)
	idsLen := len(args.IDs)
	if idsLen > 0 {
		// if asking for a specific set of ids from a given host
		// others needs to be false, or will always result in
		// an empty result set
		args.Others = false
	}
	if !args.Others {
		query.WriteString(` p.host=?`)
		queryArgs = append(queryArgs, args.Host)
		if me.AuthedExists(tlbx) {
			me := me.AuthedGet(tlbx)
			if !me.Equal(args.Host) {
				query.WriteString(` AND (p.isPublic=1 OR p.id IN (SELECT u.project FROM users u WHERE u.host=? AND u.isActive=1 AND u.id=?))`)
				queryArgs = append(queryArgs, args.Host, me)
			}
		}
	} else {
		// asking for other peoples projects which host is an active member of
		query.WriteString(` p.id IN (SELECT u.project FROM users u WHERE u.host <> ? AND u.isActive=1 AND u.id=?)`)
		queryArgs = append(queryArgs, args.Host, args.Host)
		if me.AuthedExists(tlbx) {
			me := me.AuthedGet(tlbx)
			if !me.Equal(args.Host) {
				query.WriteString(` AND (p.isPublic=1 OR p.id IN (SELECT u2.project FROM users u2 WHERE u2.host<>? AND u2.isActive=1 AND u2.id=?))`)
				queryArgs = append(queryArgs, args.Host, me)
			}
		}
	}
	if !me.AuthedExists(tlbx) {
		query.WriteString(` AND p.isPublic=1`)
	}
	if idsLen > 0 {
		query.WriteString(sql.InCondition(true, `p.id`, idsLen))
		query.WriteString(sql.OrderByField(`p.id`, idsLen))
		Is := args.IDs.ToIs()
		queryArgs = append(queryArgs, Is...)
		queryArgs = append(queryArgs, Is...)
	} else {
		query.WriteString(` AND p.isArchived=?`)
		queryArgs = append(queryArgs, args.IsArchived)
		if args.IsPublic != nil {
			query.WriteString(` AND p.isPublic=?`)
			queryArgs = append(queryArgs, *args.IsPublic)
		}
		if ptr.StringOr(args.NamePrefix, "") != "" {
			query.WriteString(` AND p.name LIKE ?`)
			*args.NamePrefix = strings.ReplaceAll(*args.NamePrefix, `\`, `\\`)
			*args.NamePrefix = strings.ReplaceAll(*args.NamePrefix, `%`, `\%`)
			queryArgs = append(queryArgs, Strf(`%s%%`, *args.NamePrefix))
		}
		if args.CreatedOnMin != nil {
			query.WriteString(` AND p.createdOn >=?`)
			queryArgs = append(queryArgs, *args.CreatedOnMin)
		}
		if args.CreatedOnMax != nil {
			query.WriteString(` AND p.createdOn <= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMax)
		}
		if args.StartOnMin != nil {
			query.WriteString(` AND p.startOn >=?`)
			queryArgs = append(queryArgs, *args.StartOnMin)
		}
		if args.StartOnMax != nil {
			query.WriteString(` AND p.startOn <=?`)
			queryArgs = append(queryArgs, *args.StartOnMax)
		}
		if args.EndOnMin != nil {
			query.WriteString(` AND p.endOn >=?`)
			queryArgs = append(queryArgs, *args.EndOnMin)
		}
		if args.EndOnMax != nil {
			query.WriteString(` AND p.endOn <=?`)
			queryArgs = append(queryArgs, *args.EndOnMax)
		}
		if args.After != nil {
			query.WriteString(Strf(` AND p.%s %s= (SELECT p.%s FROM projects p WHERE p.host=? AND p.id=?) AND p.id <> ?`, args.Sort, sql.GtLtSymbol(*args.Asc), args.Sort))
			queryArgs = append(queryArgs, args.Host, *args.After, *args.After)
			if args.Sort != cnsts.SortCreatedOn {
				query.WriteString(Strf(` AND p.createdOn %s (SELECT p.createdOn FROM projects p WHERE p.host=? AND p.id=?)`, sql.GtLtSymbol(*args.Asc)))
				queryArgs = append(queryArgs, args.Host, *args.After)
			}
		}
		createdOnSecondarySort := ", p.id"
		if args.Sort != cnsts.SortCreatedOn {
			createdOnSecondarySort = ", p.createdOn, p.id"
		}
		query.WriteString(sql.OrderLimit100("p."+string(args.Sort)+createdOnSecondarySort, *args.Asc, args.Limit))
	}
	PanicOn(srv.Data().Query(func(rows isql.Rows) {
		iLimit := int(args.Limit)
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == iLimit {
				res.More = true
				break
			}
			p := &project.Project{}
			PanicOn(rows.Scan(&p.Host, &p.ID, &p.IsArchived, &p.Name, &p.CreatedOn, &p.CurrencyCode, &p.HoursPerDay, &p.DaysPerWeek, &p.StartOn, &p.EndOn, &p.IsPublic, &p.FileLimit, &p.Parent, &p.FirstChild, &p.NextSib, &p.User, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedOn, &p.TimeEst, &p.TimeInc, &p.TimeSubMin, &p.TimeSubEst, &p.TimeSubInc, &p.CostEst, &p.CostInc, &p.CostSubEst, &p.CostSubInc, &p.FileN, &p.FileSize, &p.FileSubN, &p.FileSubSize, &p.ChildN, &p.DescN, &p.IsParallel))
			res.Set = append(res.Set, p)
		}
	}, query.String(), queryArgs...))
	return res
}

func getLatestPublicProjects(tlbx app.Tlbx) *project.GetRes {
	limit := 100
	srv := service.Get(tlbx)
	res := &project.GetRes{
		Set: make([]*project.Project, 0, limit),
	}
	query := bytes.NewBufferString(projects_select_columns)
	query.WriteString(` p.isPublic=1 ORDER BY p.createdOn DESC`)
	PanicOn(srv.Data().Query(func(rows isql.Rows) {
		for rows.Next() {
			p := &project.Project{}
			PanicOn(rows.Scan(&p.Host, &p.ID, &p.IsArchived, &p.Name, &p.CreatedOn, &p.CurrencyCode, &p.HoursPerDay, &p.DaysPerWeek, &p.StartOn, &p.EndOn, &p.IsPublic, &p.FileLimit, &p.Parent, &p.FirstChild, &p.NextSib, &p.User, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedOn, &p.TimeEst, &p.TimeInc, &p.TimeSubMin, &p.TimeSubEst, &p.TimeSubInc, &p.CostEst, &p.CostInc, &p.CostSubEst, &p.CostSubInc, &p.FileN, &p.FileSize, &p.FileSubN, &p.FileSubSize, &p.ChildN, &p.DescN, &p.IsParallel))
			res.Set = append(res.Set, p)
		}
	}, query.String()))
	return res
}

func getUsers(tlbx app.Tlbx, args *project.GetUsers) *project.GetUsersRes {
	validate.MaxIDs(tlbx, "ids", args.IDs, 100)
	app.BadReqIf(args.HandlePrefix != nil && StrLen(*args.HandlePrefix) >= 15, "handlePrefix must be < 15 chars long")
	tx := service.Get(tlbx).Data().BeginRead()
	defer tx.Rollback()
	epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
	limit := sql.Limit100(args.Limit)
	res := &project.GetUsersRes{
		Set: make([]*project.User, 0, limit),
	}
	query := bytes.NewBufferString(`WITH`)
	queryArgs := make([]interface{}, 0, 14+len(args.IDs))
	if args.After != nil {
		query.WriteString(` after (id, role, handle) AS (SELECT u.id, u.role, u.handle FROM users u WHERE u.host=? AND u.project=? AND u.id=?),`)
		queryArgs = append(queryArgs, args.Host, args.Project, *args.After)
	}
	query.WriteString(` selector (host, project, id) AS (SELECT u.host, u.project, u.id FROM users u WHERE u.host=? AND u.project=?`)
	queryArgs = append(queryArgs, args.Host, args.Project)
	idsLen := len(args.IDs)
	Is := args.IDs.ToIs()
	if idsLen > 0 {
		query.WriteString(sql.InCondition(true, `u.id`, idsLen))
		queryArgs = append(queryArgs, Is...)
	} else {
		query.WriteString(` AND u.isActive=1`)
		if ptr.StringOr(args.HandlePrefix, "") != "" {
			query.WriteString(` AND u.handle LIKE ?`)
			queryArgs = append(queryArgs, Strf(`%s%%`, *args.HandlePrefix))
		}
		if args.Role != nil {
			query.WriteString(` AND u.role=?`)
			queryArgs = append(queryArgs, *args.Role)
		}
		if args.After != nil {
			query.WriteString(` AND (((u.id=?) < ((SELECT id FROM after)=?)) OR (u.role > (SELECT role FROM after) OR (u.role = (SELECT role FROM after) AND u.handle > (SELECT handle FROM after))))`)
			queryArgs = append(queryArgs, args.Host, args.Host)
		}
		query.WriteString(Strf(` ORDER BY (u.id=?) DESC, role ASC, handle ASC LIMIT %d`, limit))
		queryArgs = append(queryArgs, args.Host)
	}
	query.WriteString(`) SELECT u.id, u.handle, u.alias, u.hasAvatar, u.isActive, u.role, t.timeEst, vi.timeInc, t.costEst, vi.costInc, f.fileN, f.fileSize, t.taskN FROM users u JOIN (SELECT s.id AS id, COALESCE(SUM(t.timeEst), 0) AS timeEst, COALESCE(SUM(t.costEst), 0) AS costEst, COALESCE(COUNT(t.id), 0) AS taskN FROM selector s LEFT JOIN tasks t ON (t.host=s.host AND t.project=s.project AND t.user=s.id) GROUP BY s.id) t ON t.id = u.id JOIN (SELECT  s.id AS id, COALESCE(SUM(CASE vi.type WHEN 'time' THEN vi.inc ELSE 0 END), 0) AS timeInc, COALESCE(SUM(CASE vi.type WHEN 'cost' THEN vi.inc ELSE 0 END), 0) AS costInc FROM selector s LEFT JOIN vitems vi ON (vi.host=s.host AND vi.project=s.project AND vi.createdBy=s.id) GROUP BY s.id) vi ON vi.id = u.id JOIN (SELECT s.id AS id, COALESCE(COUNT(f.id), 0) AS fileN, COALESCE(SUM(f.size), 0) AS fileSize FROM selector s LEFT JOIN files f ON (f.host=s.host AND f.project=s.project AND f.createdBy=s.id) GROUP BY s.id) f ON f.id = u.id JOIN selector s ON u.host=s.host AND u.project=s.project AND u.id = s.id`)

	if idsLen > 0 {
		query.WriteString(sql.OrderByField(`u.id`, idsLen))
		queryArgs = append(queryArgs, Is...)
	} else {
		query.WriteString(` ORDER BY (u.id=?) DESC, role ASC, handle ASC`)
		queryArgs = append(queryArgs, args.Host)
	}
	Println(query.String())
	Println(queryArgs...)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == int(limit) {
				res.More = true
				break
			}
			u := &project.User{}
			PanicOn(rows.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar, &u.IsActive, &u.Role, &u.TimeEst, &u.TimeInc, &u.CostEst, &u.CostInc, &u.FileN, &u.FileSize, &u.TaskN))
			res.Set = append(res.Set, u)
		}
	}, query.String(), queryArgs...))
	tx.Commit()
	return res
}

var (
	projects_select_columns = `SELECT p.host, p.id, p.isArchived, p.name, p.createdOn, p.currencyCode, p.hoursPerDay, p.daysPerWeek, p.startOn, p.endOn, p.isPublic, p.fileLimit, t.parent, t.firstChild, t.nextSib, t.user, t.name, t.description, t.createdBy, t.createdOn, t.timeEst, t.timeInc, t.timeSubMin, t.timeSubEst, t.timeSubInc, t.costEst, t.costInc, t.costSubEst, t.costSubInc, t.fileN, t.fileSize, t.fileSubN, t.fileSubSize, t.childN, t.descN, t.isParallel FROM projects p JOIN tasks t ON (t.host=p.host AND t.project=p.id AND t.id=p.id) WHERE`
)

func validateCurrencyCode(tlbx app.Tlbx, code string) {
	app.BadReqIf(!currencies[code], "invalid currency code")
}

var currencies = map[string]bool{
	"AED": true,
	"AFN": true,
	"ALL": true,
	"AMD": true,
	"ANG": true,
	"AOA": true,
	"ARS": true,
	"AUD": true,
	"AWG": true,
	"AZN": true,
	"BAM": true,
	"BBD": true,
	"BDT": true,
	"BGN": true,
	"BHD": true,
	"BIF": true,
	"BMD": true,
	"BND": true,
	"BOB": true,
	"BOV": true,
	"BRL": true,
	"BSD": true,
	"BTN": true,
	"BWP": true,
	"BYN": true,
	"BZD": true,
	"CAD": true,
	"CDF": true,
	"CHE": true,
	"CHF": true,
	"CHW": true,
	"CLF": true,
	"CLP": true,
	"CNY": true,
	"COP": true,
	"COU": true,
	"CRC": true,
	"CUC": true,
	"CUP": true,
	"CVE": true,
	"CZK": true,
	"DJF": true,
	"DKK": true,
	"DOP": true,
	"DZD": true,
	"EGP": true,
	"ERN": true,
	"ETB": true,
	"EUR": true,
	"FJD": true,
	"FKP": true,
	"GBP": true,
	"GEL": true,
	"GHS": true,
	"GIP": true,
	"GMD": true,
	"GNF": true,
	"GTQ": true,
	"GYD": true,
	"HKD": true,
	"HNL": true,
	"HRK": true,
	"HTG": true,
	"HUF": true,
	"IDR": true,
	"ILS": true,
	"INR": true,
	"IQD": true,
	"IRR": true,
	"ISK": true,
	"JMD": true,
	"JOD": true,
	"JPY": true,
	"KES": true,
	"KGS": true,
	"KHR": true,
	"KMF": true,
	"KPW": true,
	"KRW": true,
	"KWD": true,
	"KYD": true,
	"KZT": true,
	"LAK": true,
	"LBP": true,
	"LKR": true,
	"LRD": true,
	"LSL": true,
	"LYD": true,
	"MAD": true,
	"MDL": true,
	"MGA": true,
	"MKD": true,
	"MMK": true,
	"MNT": true,
	"MOP": true,
	"MRU": true,
	"MUR": true,
	"MVR": true,
	"MWK": true,
	"MXN": true,
	"MXV": true,
	"MYR": true,
	"MZN": true,
	"NAD": true,
	"NGN": true,
	"NIO": true,
	"NOK": true,
	"NPR": true,
	"NZD": true,
	"OMR": true,
	"PAB": true,
	"PEN": true,
	"PGK": true,
	"PHP": true,
	"PKR": true,
	"PLN": true,
	"PYG": true,
	"QAR": true,
	"RON": true,
	"RSD": true,
	"RUB": true,
	"RWF": true,
	"SAR": true,
	"SBD": true,
	"SCR": true,
	"SDG": true,
	"SEK": true,
	"SGD": true,
	"SHP": true,
	"SLL": true,
	"SOS": true,
	"SRD": true,
	"SSP": true,
	"STN": true,
	"SVC": true,
	"SYP": true,
	"SZL": true,
	"THB": true,
	"TJS": true,
	"TMT": true,
	"TND": true,
	"TOP": true,
	"TRY": true,
	"TTD": true,
	"TWD": true,
	"TZS": true,
	"UAH": true,
	"UGX": true,
	"USD": true,
	"USN": true,
	"UYI": true,
	"UYU": true,
	"UYW": true,
	"UZS": true,
	"VES": true,
	"VND": true,
	"VUV": true,
	"WST": true,
	"XAF": true,
	"XAG": true,
	"XAU": true,
	"XBA": true,
	"XBB": true,
	"XBC": true,
	"XBD": true,
	"XCD": true,
	"XDR": true,
	"XOF": true,
	"XPD": true,
	"XPF": true,
	"XPT": true,
	"XSU": true,
	"XTS": true,
	"XUA": true,
	"XXX": true,
	"YER": true,
	"ZAR": true,
	"ZMW": true,
	"ZWL": true,
}
