package validate

import (
	"regexp"

	"github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

func Str(name, str string, tlbx app.Tlbx, minLen, maxLen int, regexs ...*regexp.Regexp) {
	app.BadReqIf(minLen > 0 && core.StrLen(str) < minLen, "%s does not satisfy min len %d", name, minLen)
	app.BadReqIf(maxLen > 0 && core.StrLen(str) > maxLen, "%s does not satisfy max len %d", name, maxLen)
	for _, re := range regexs {
		app.BadReqIf(!re.MatchString(str), "%s does not satisfy regexp %s", name, re)
	}
}

func MaxIDs(tlbx app.Tlbx, name string, ids []core.ID, max int) {
	app.BadReqIf(len(ids) > max, "%s must not contain more than %d ids", name, max)
}
