package itemtest

import (
	"testing"
	"time"

	"github.com/0xor1/wtf/cmd/todo/pkg/item"
	"github.com/0xor1/wtf/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/wtf/cmd/todo/pkg/list"
	"github.com/0xor1/wtf/cmd/todo/pkg/list/listeps"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/field"
	"github.com/0xor1/wtf/pkg/ptr"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(append(listeps.Eps, itemeps.Eps...), listeps.OnDelete)
	defer r.CleanUp()

	testList1 := (&list.Create{
		Name: "Test list 1",
	}).MustDo(r.Ali().Client())

	name1 := "Test item 1"
	testItem1 := (&item.Create{
		List: testList1.ID,
		Name: name1,
	}).MustDo(r.Ali().Client())
	a.Equal(name1, testItem1.Name)

	name2 := "Test item 2"
	testItem2 := (&item.Create{
		List: testList1.ID,
		Name: name2,
	}).MustDo(r.Ali().Client())
	a.Equal(name2, testItem2.Name)

	one := (&item.One{
		List: testList1.ID,
		ID:   testItem1.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1, one)

	nilOne := (&item.One{
		List: testList1.ID,
		ID:   app.ExampleID(),
	}).MustDo(r.Ali().Client())
	a.Nil(nilOne)

	get := (&item.Get{
		List: testList1.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1, get.Set[0])
	a.Equal(testItem2, get.Set[1])
	a.False(get.More)

	get = (&item.Get{
		List: testList1.ID,
		IDs:  IDs{testItem2.ID, testItem1.ID},
	}).MustDo(r.Ali().Client())
	a.Equal(testItem2, get.Set[0])
	a.Equal(testItem1, get.Set[1])
	a.False(get.More)

	get = (&item.Get{
		List:           testList1.ID,
		NameStartsWith: ptr.String("Test i"),
		CreatedOnMin:   ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax:   ptr.Time(Now()),
		Asc:            ptr.Bool(false),
		Limit:          ptr.Int(2),
	}).MustDo(r.Ali().Client())
	a.Equal(testItem2, get.Set[0])
	a.Equal(testItem1, get.Set[1])
	a.False(get.More)

	get = (&item.Get{
		List:           testList1.ID,
		NameStartsWith: ptr.String("Test i"),
		CreatedOnMin:   ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax:   ptr.Time(Now()),
		After:          ptr.ID(testItem1.ID),
		Sort:           item.SortName,
		Asc:            ptr.Bool(true),
		Limit:          ptr.Int(2),
	}).MustDo(r.Ali().Client())
	a.Equal(testItem2, get.Set[0])
	a.False(get.More)

	get = (&item.Get{
		List:           testList1.ID,
		NameStartsWith: ptr.String("Test i"),
		CreatedOnMin:   ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax:   ptr.Time(Now()),
		Asc:            ptr.Bool(true),
		Limit:          ptr.Int(1),
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1, get.Set[0])
	a.True(get.More)

	newName := "New Name"
	updatedItem := (&item.Update{
		List:     testList1.ID,
		ID:       testItem1.ID,
		Name:     &field.String{Val: newName},
		Complete: &field.Bool{Val: true},
	}).MustDo(r.Ali().Client())
	testItem1.Name = newName
	a.Equal(testItem1.Name, updatedItem.Name)
	a.NotNil(updatedItem.CompletedOn)

	(&item.Delete{}).MustDo(r.Ali().Client())
	(&item.Delete{IDs: IDs{testItem1.ID}}).MustDo(r.Ali().Client())
}
