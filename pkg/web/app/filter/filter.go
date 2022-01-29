package filter

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Base struct {
	IDs        IDs      `json:"ids,omitempty"`
	After      *ID      `json:"after,omitempty"`
	Sort       string   `json:"sort,omitempty"`
	Asc        *bool    `json:"asc,omitempty"`
	Limit      uint16   `json:"limit,omitempty"`
	MaxLimit   uint16   `json:"-"`
	ValidSorts []string `json:"-"`
}

func (b *Base) MustBeValid() {
	app.BadReqIf(len(b.IDs) > int(b.MaxLimit), "%d ids supplied, max limit %d", len(b.IDs), b.MaxLimit)
	app.BadReqIf(b.Limit > b.MaxLimit, "limit of %d is larger than max limit %d", b.Limit, b.MaxLimit)
	if len(b.ValidSorts) == 0 {
		b.Sort = ""
	} else if b.Sort == "" {
		// no sort specified, use default sort
		b.Sort = b.ValidSorts[0]
	} else {
		matchFound := false
		lowSort := StrLower(b.Sort)
		for _, s := range b.ValidSorts {
			if StrLower(s) == lowSort {
				// if sort does match case insensitive
				// then reassign just in case it wasnt
				// case sensitive matching
				b.Sort = s
				matchFound = true
				break
			}
		}
		app.BadReqIf(!matchFound, "invalid sort %s, valid sort options %v", b.Sort, b.ValidSorts)
	}
}

func Defs(asc bool, limit, maxLimit uint16, validSorts ...string) Base {
	sort := ""
	if len(validSorts) > 0 {
		sort = validSorts[0]
	}
	return Base{
		Sort:       sort,
		Asc:        ptr.Bool(asc),
		Limit:      limit,
		MaxLimit:   maxLimit,
		ValidSorts: validSorts,
	}
}

func DefsAsc100(validSorts ...string) Base {
	return Defs(true, 100, 100, validSorts...)
}
