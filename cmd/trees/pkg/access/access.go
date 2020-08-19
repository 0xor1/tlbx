package access

import (
	"net/http"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
)

func ProjectCheck(tlbx app.Tlbx, host, project ID, role cnsts.Role) {
	iExist := me.Exists(tlbx)
	if iExist && me.Get(tlbx).Equal(host) {
		return
	}

	srv := service.Get(tlbx)
	if !iExist || role == cnsts.RoleReader {
		row := srv.Data().QueryRow(`SELECT isPublic FROM projects WHERE host=? AND id=?`, host, project)
		isPublic := false
		sql.PanicIfIsntNoRows(row.Scan(&isPublic))
		if isPublic {
			return
		}
		// at this point project isnt public so if no active session return 403
		app.ReturnIf(!iExist, http.StatusForbidden, "")
	}

	row := srv.Data().QueryRow(`SELECT TRUE FROM projectUsers WHERE host=? AND project=? AND id=? AND role<=?`, host, project, me.Get(tlbx), role)
	hasAccess := false
	sql.PanicIfIsntNoRows(row.Scan(&hasAccess))
	app.ReturnIf(!hasAccess, http.StatusForbidden, "")
}
