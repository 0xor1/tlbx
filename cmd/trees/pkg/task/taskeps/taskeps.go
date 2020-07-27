package taskeps

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/user"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new task",
			Path:         (&task.Create{}).Path(),
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
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return nil
			},
		},
	}
	aliasMaxLen = 50
)

func OnActivate(tlbx app.Tlbx, me *user.User) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	// _, err := tx.Exec(`INSERT INTO accounts WHERE id=?`, me)
	// PanicOn(err)
	tx.Commit()
}

func OnDelete(tlbx app.Tlbx, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	// _, err := tx.Exec(`DELETE FROM accounts WHERE id=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM times WHERE account=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM tasks WHERE account=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM projects WHERE account=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM projects WHERE account=?`, me)
	// PanicOn(err)
	tx.Commit()
}
