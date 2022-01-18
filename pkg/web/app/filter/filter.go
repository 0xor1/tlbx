package filter

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type rawJsonBase struct {
	IDs   IDs    `json:"ids,omitempty"`
	After *ID    `json:"after,omitempty"`
	Sort  string `json:"sort,omitempty"`
	Asc   *bool  `json:"asc,omitempty"`
	Limit uint16 `json:"limit,omitempty"`
}

type Base struct {
	IDs        IDs      `json:"ids,omitempty"`
	After      *ID      `json:"after,omitempty"`
	Sort       string   `json:"sort,omitempty"`
	Asc        *bool    `json:"asc,omitempty"`
	Limit      uint16   `json:"limit,omitempty"`
	MaxIDs     uint16   `json:"-"`
	MaxLimit   uint16   `json:"-"`
	ValidSorts []string `json:"-"`
}

func (b *Base) UnmarshalJSON(data []byte) error {
	rjs := &rawJsonBase{
		IDs:   b.IDs,
		After: b.After,
		Sort:  b.Sort,
		Asc:   b.Asc,
		Limit: b.Limit,
	}
	err := json.Unmarshal(data, &rjs)
	if err != nil {
		return err
	}
	*b = Base{
		IDs:        rjs.IDs,
		After:      rjs.After,
		Sort:       rjs.Sort,
		Asc:        rjs.Asc,
		Limit:      rjs.Limit,
		MaxIDs:     b.MaxIDs,
		MaxLimit:   b.MaxLimit,
		ValidSorts: b.ValidSorts,
	}
	b.mustBeValid()
	return nil
}

func (b *Base) mustBeValid() {
	app.BadReqIf(len(b.IDs) > int(b.MaxIDs), "%d ids supplied, max ids %d", len(b.IDs), b.MaxIDs)
	app.BadReqIf(b.Limit > b.MaxLimit, "limit of %d is larger than max limit %d", b.Limit, b.MaxLimit)
	if b.Sort == "" && len(b.ValidSorts) > 0 {
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

func Defs(asc bool, limit, maxIDs, maxLimit uint16, validSorts ...string) Base {
	return Base{
		Sort:       validSorts[0],
		Asc:        ptr.Bool(asc),
		Limit:      limit,
		MaxIDs:     maxIDs,
		MaxLimit:   maxLimit,
		ValidSorts: validSorts,
	}
}

func DefsAsc100(validSorts ...string) Base {
	return Defs(true, 100, 100, 100, validSorts...)
}
