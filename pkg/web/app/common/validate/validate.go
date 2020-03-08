package validate

import (
	"net/http"
	"regexp"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

func Str(name, str string, tlbx app.Toolbox, minLen, maxLen int, regexs ...*regexp.Regexp) {
	tlbx.ReturnMsgIf(minLen > 0 && StrLen(str) < minLen, http.StatusBadRequest, "%s does not satisfy min len %d", name, minLen)
	tlbx.ReturnMsgIf(maxLen > 0 && StrLen(str) > maxLen, http.StatusBadRequest, "%s does not satisfy max len %d", name, maxLen)
	for _, re := range regexs {
		tlbx.ReturnMsgIf(!re.MatchString(str), http.StatusBadRequest, "%s does not satisfy regexp %s", name, re)
	}
}
