package validate

import (
	"regexp"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

func Str(name, str string, tlbx app.Tlbx, minLen, maxLen int, regexs ...*regexp.Regexp) {
	tlbx.BadReqIf(minLen > 0 && StrLen(str) < minLen, "%s does not satisfy min len %d", name, minLen)
	tlbx.BadReqIf(maxLen > 0 && StrLen(str) > maxLen, "%s does not satisfy max len %d", name, maxLen)
	for _, re := range regexs {
		tlbx.BadReqIf(!re.MatchString(str), "%s does not satisfy regexp %s", name, re)
	}
}

func MaxIDs(tlbx app.Tlbx, name string, ids []ID, max int) {
	tlbx.BadReqIf(len(ids) > max, "%s must not contain more than %d ids", name, max)
}
