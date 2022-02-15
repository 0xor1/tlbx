package json

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"

	"github.com/stretchr/testify/assert"
)

func Test_Wrappers(t *testing.T) {
	a := assert.New(t)
	bs := MustMarshal(nil)
	a.Equal("null", string(bs))
	bs = MustMarshalIndent(nil, "", "    ")
	a.Equal("null", string(bs))
	var v interface{}
	MustUnmarshal(bs, &v)
	a.Nil(v)
	MustUnmarshalReader(bytes.NewReader(bs), &v)
	a.Nil(v)
}

func Test_SplitPath(t *testing.T) {
	a := assert.New(t)

	path, err := SplitPath("one.2.three")
	a.Nil(err)
	a.Equal([]interface{}{"one", 2, "three"}, path)

	path, err = SplitPath("one.2.three.")
	a.Equal(err, emptyPathPartErr)
	a.Nil(path)

	path, err = SplitPath("one.99999999999999999999999999999999999999999999999999999999.three")
	a.NotNil(err)
	a.Nil(path)

	MustSplitPath("one.2.three")
}

func Test_New(t *testing.T) {
	a := assert.New(t)

	js, err := New()
	a.Nil(err, "err is nil")

	str, err := js.ToString()
	a.Nil(err, "err is nil")
	a.Equal("{}", str, "str is an empty json object string")
}

func Test_MustNew(t *testing.T) {
	MustNew()
}

func Test_FromString(t *testing.T) {
	a := assert.New(t)

	src := `{"foo":[1,true,"bar"]}`
	js, err := FromString(src)
	a.Nil(err, "err is nil")

	str, err := js.ToString()
	a.Nil(err, "err is nil")
	a.Equal(src, str, "str is an empty json object string")
}

func Test_MustFromString(t *testing.T) {
	MustFromString(`{}`)
}

func Test_MustFromBytes(t *testing.T) {
	MustFromBytes([]byte(`{}`))
}

func Test_MustToBytes(t *testing.T) {
	a := assert.New(t)
	a.Equal(MustNew().MustToBytes(), []byte(`{}`))
}

func Test_MustToString(t *testing.T) {
	a := assert.New(t)
	a.Equal(MustNew().MustToString(), `{}`)
}

func Test_MustToPrettyBytes(t *testing.T) {
	a := assert.New(t)
	a.Equal(MustNew().MustToPrettyBytes(), []byte(`{}`))
}

func Test_FromInterface(t *testing.T) {
	a := assert.New(t)

	obj, err := New()
	a.Nil(err, "err is nil")
	i, err := obj.Interface()
	a.Nil(err, "err is nil")
	obj2 := FromInterface(i)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal("{}", str, "str is an empty json object string")
	str2, err := obj2.ToString()
	a.Nil(err, "err is nil")
	a.Equal("{}", str2, "str2 is an empty json object string")
}

func Test_FromFile(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"one":1,"foo":"bar"}`)
	a.Nil(err, "err is nil")

	wd, _ := os.Getwd()
	file := filepath.Join(wd, "test.json")
	err = obj.ToFile(file, os.ModePerm)
	a.Nil(err, "err is nil")
	a.Nil(os.Remove(file))
	obj.MustToFile(file, os.ModePerm)

	obj2, err := FromFile(file)
	a.Nil(err, "err is nil")

	MustFromFile(file)

	a.Nil(os.Remove(file))

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	str2, err := obj2.ToString()
	a.Nil(err, "err is nil")
	a.Nil(err, "err is nil")
	a.Equal(str, str2, "both strings are equal")
}

func Test_FromFile_error(t *testing.T) {
	a := assert.New(t)

	wd, _ := os.Getwd()
	file := filepath.Join(wd, "test.json")
	obj, err := FromFile(file)
	a.Nil(obj, "obj is nil")
	a.True(os.IsNotExist(err), "err is a not exists error")
}

func Test_FromReader_Nil(t *testing.T) {
	a := assert.New(t)

	obj, err := FromReader(nil)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal("null", str, "str is an empty json object string")

	MustFromReader(nil)
}

func Test_FromReadCloser_Nil(t *testing.T) {
	a := assert.New(t)

	obj, err := FromReadCloser(nil)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal("null", str, "str is an empty json object string")

	MustFromReadCloser(nil)
}

func Test_UnmarshalJSON(t *testing.T) {
	a := assert.New(t)

	obj := &Json{}
	err := obj.UnmarshalJSON([]byte("{}"))
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal("{}", str, "str is an empty json object")
}

func Test_UnmarshalJSON_WithMalformedJson(t *testing.T) {
	a := assert.New(t)

	obj := &Json{}
	err := obj.UnmarshalJSON([]byte("{"))
	a.NotNil(err, "err is not nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal("null", str, "str is json null value")
}

func Test_ToPrettyString(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":1,"b":2}`)
	a.Nil(err, "err is nil")

	str, err := obj.ToPrettyString()
	a.Nil(err, "err is nil")
	a.Equal("{\n  \"a\": 1,\n  \"b\": 2\n}", str, "str is indented json object")

	obj.MustToPrettyString()
}

func Test_ToReader(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":1,"b":2}`)
	a.Nil(err, "err is nil")

	reader, err := obj.ToReader()
	a.Nil(err, "err is nil")
	obj, err = FromReader(reader)
	a.Nil(err, "err is nil")
	str, err := obj.ToString()
	a.Equal(`{"a":1,"b":2}`, str, "str is the json object")

	obj.MustToReader()
}

func Test_Exists(t *testing.T) {
	a := assert.New(t)

	a.True(MustFromString(`{"v": null}`).Exists("v"))
	a.False(MustFromString(`{"v": null}`).Exists("x"))
}

func Test_Get(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	obj2, err := obj.Get(MustParseKey("a"), 1, "b", 2, "c")
	a.Nil(err, "err is nil")
	obj2 = obj.MustGet("a", 1, "b", 2, "c")

	str := obj2.StringOrDefault("")
	a.Equal("got it!", str, "str is correct value")

	obj2.MustString()
}

func Test_Get_WithMissingMapKey(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	obj, pathErr := obj.Get("a", 1, "b", 2, "d")
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a", 1, "b", 2}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"d"}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"c":"got it!"}`, str, "str is correct value")
}

func Test_Get_WithInappropriateMapKey(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	obj, pathErr := obj.Get("a", 1, "b", "c")
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a", 1, "b"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"c"}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`[[],{},{"c":"got it!"}]`, str, "str is correct value")
}

func Test_Get_WithOutOfBoundsSliceIndex(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	obj, pathErr := obj.Get("a", 1, "b", 0, 0)
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a", 1, "b", 0}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{0}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`[]`, str, "str is correct value")
}

func Test_Get_WithInappropriateSliceIndex(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	obj, pathErr := obj.Get("a", 1, 0, "b")
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a", 1}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{0, "b"}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"b":[[],{},{"c":"got it!"}]}`, str, "str is correct value")
}

func Test_Get_WithInappropriatePathValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	obj, pathErr := obj.Get("a", 1, true)
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a", 1}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{true}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"b":[[],{},{"c":"got it!"}]}`, str, "str is correct value")
}

func Test_Set_WithMapKey(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	err = obj.Set("a", 1, "b", 2, "d", "set it!")
	a.Nil(err, "err is nil")
	obj.MustSet("a", 1, "b", 2, "d", "set it!")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":[{},{"b":[[],{},{"c":"got it!","d":"set it!"}]}]}`, str, "str is correct value")

	obj.Set("a", nil)
}

func Test_Set_Empty(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	err = obj.Set(nil)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal("null", str, "str is correct value")
}

func Test_Set_NoPath(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"got it!"}]}]}`)
	a.Nil(err, "err is nil")

	err = obj.Set(true)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal("true", str, "str is correct value")
}

func Test_Set_NoPathAndValue(t *testing.T) {
	a := assert.New(t)

	obj, err := New()
	a.Nil(err, "err is nil")

	err = obj.Set()
	a.NotNil(err, "err is not nil")
}

func Test_Set_NestedNonExistantMaps(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{}`)
	a.Nil(err, "err is nil")

	err = obj.Set("a", "b", "c", true)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":{"b":{"c":true}}}`, str, "str is correct value")
}

func Test_Set_WithInappropriateMapKey(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":true}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Set("a", "b", true)
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":true}`, str, "str is correct value")
}

func Test_Set_WithSliceIndex(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[null]}`)
	a.Nil(err, "err is nil")

	err = obj.Set("a", 0, true)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":[true]}`, str, "str is correct value")
}

func Test_Set_WithInappropriateSliceIndex(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[]}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Set("a", 0, true)
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{0}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":[]}`, str, "str is correct value")
}

func Test_Set_WithInappropriatePathValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[]}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Set("a", true, true)
	a.NotNil(pathErr, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{true}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":[]}`, str, "str is correct value")
}

func Test_Del_WithMapKey(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"delete me!"}]}]}`)
	a.Nil(err, "err is nil")

	err = obj.Del("a", 1, "b", 2, "c")
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":[{},{"b":[[],{},{}]}]}`, str, "str is correct value")

	obj.MustDel("a", 0)
}

func Test_Del_WithSliceIndex(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[{},{"b":[[],{},{"c":"delete me!"}]}]}`)
	a.Nil(err, "err is nil")

	err = obj.Del("a", 1, "b", 2)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":[{},{"b":[[],{}]}]}`, str, "str is correct value")
}

func Test_Del_WithRootSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	err = obj.Del(1)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`[0,2]`, str, "str is correct value")
}

func Test_Del_WithNestedSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,[0,1,2]]`)
	a.Nil(err, "err is nil")

	err = obj.Del(2, 1)
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`[0,1,[0,2]]`, str, "str is correct value")
}

func Test_Del_WithEmptyPath(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":{"b":{"c":"delete me!"}}}`)
	a.Nil(err, "err is nil")

	err = obj.Del()
	a.Nil(err, "err is nil")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`null`, str, "str is correct value")
}

func Test_Del_WithIncorrectPathValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":{"b":{"c":"delete me!"}}}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Del("a", "c", "b")
	a.NotNil(pathErr, "err is nil")
	a.Equal([]interface{}{"a"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"c", "b"}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":{"b":{"c":"delete me!"}}}`, str, "str is correct value")
}

func Test_Del_WithInappropriateLastMapKey(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":{"b":{"c":["delete me!"]}}}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Del("a", "b", "c", "d")
	a.NotNil(pathErr, "err is nil")
	a.Equal([]interface{}{"a", "b", "c"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"d"}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":{"b":{"c":["delete me!"]}}}`, str, "str is correct value")
}

func Test_Del_WithInappropriateLastSliceIndex(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":{"b":{"c":"delete me!"}}}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Del("a", "b", 1)
	a.NotNil(pathErr, "err is nil")
	a.Equal([]interface{}{"a", "b"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{1}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":{"b":{"c":"delete me!"}}}`, str, "str is correct value")
}

func Test_Del_WithOutOfBoundsLastSliceIndex(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":{"b":{"c":["delete me!"]}}}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Del("a", "b", "c", 1)
	a.NotNil(pathErr, "err is nil")
	a.Equal([]interface{}{"a", "b", "c"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{1}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":{"b":{"c":["delete me!"]}}}`, str, "str is correct value")
}

func Test_Del_WithInappropriateLastPathValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":{"b":{"c":"delete me!"}}}`)
	a.Nil(err, "err is nil")

	pathErr := obj.Del("a", "b", true)
	a.NotNil(pathErr, "err is nil")
	a.Equal([]interface{}{"a", "b"}, ToError(pathErr).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{true}, ToError(pathErr).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a b] missing: [true]", ToError(pathErr).Value().(*jsonPathError).Error(), "error message is correct")

	str, err := obj.ToString()
	a.Nil(err, "err is nil")
	a.Equal(`{"a":{"b":{"c":"delete me!"}}}`, str, "str is correct value")
}

func Test_Interface(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": true}`)
	a.Nil(err, "err is nil")

	val, err := obj.Interface()
	val = obj.MustInterface()
	a.Nil(err, "err is nil")
	a.Equal(map[string]interface{}{"a": true}, val, "val is correct")

	val = obj.InterfaceOrDefault("a", false)
	a.Equal(true, val)
	val = obj.InterfaceOrDefault("b", false)
	a.Equal(false, val)
}

func Test_Map(t *testing.T) {
	a := assert.New(t)

	val := map[string]interface{}{}

	a.Equal(val, (&Json{nil}).MustMap())
}

func Test_Map_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":true}`)
	a.Nil(err, "err is nil")

	val := obj.MustMap()
	a.Equal(map[string]interface{}{"a": true}, val)
	val, err = obj.Map("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Nil(val, "val is correct")
}

func Test_MapOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[{"a":true}]`)
	a.Nil(err, "err is nil")

	val := obj.MapOrDefault(0, nil)
	a.Equal(map[string]interface{}{"a": true}, val, "val is correct")
}

func Test_MapString(t *testing.T) {
	a := assert.New(t)

	val := map[string]string{}

	a.Equal(val, (&Json{nil}).MustMapString())
	a.Equal(val, (&Json{val}).MustMapString())
}

func Test_MapString_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":true}`)
	a.Nil(err, "err is nil")

	val, err := obj.MapString("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Nil(val, "val is correct")
}

func Test_MapString_ValueTypeError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":true}`)
	a.Nil(err, "err is nil")

	val, err := obj.MapString()
	a.NotNil(err, "err is not nil")
	a.Equal(invalidTypeErr, err)
	a.Nil(val, "val is correct")
}

func Test_MapString_MapTypeError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`["a",true]`)
	a.Nil(err, "err is nil")

	val, err := obj.MapString()
	a.NotNil(err, "err is not nil")
	a.Equal(invalidTypeErr, err)
	a.Nil(val, "val is correct")
}

func Test_MustMapString(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[{"a":"b"}]`)
	a.Nil(err, "err is nil")

	val := obj.MustMapString(0)
	a.Equal(map[string]string{"a": "b"}, val, "val is correct")
}

func Test_MapStringOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[{"a":"b"}]`)
	a.Nil(err, "err is nil")

	val := obj.MapStringOrDefault(0, nil)
	a.Equal(map[string]string{"a": "b"}, val, "val is correct")

	obj, err = FromString(`[{"a":"b"}]`)
	a.Nil(err, "err is nil")

	def := map[string]string{"c": "d"}
	val = obj.MapStringOrDefault(def)
	a.Equal(def, val, "val is correct")
}

func Test_MustMap_DefaultValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[]`)
	a.Nil(err, "err is nil")

	val := obj.MapOrDefault(map[string]interface{}{"a": true})
	a.Equal(map[string]interface{}{"a": true}, val, "val is correct")
}

func Test_Slice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[true,false,true]}`)
	a.Nil(err, "err is nil")

	val, err := obj.Slice("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Nil(val, "val is nil")
}

func Test_SliceOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[true,false,true]}`)
	a.Nil(err, "err is nil")

	val := obj.SliceOrDefault("a", []interface{}{})
	a.Equal([]interface{}{true, false, true}, val, "val is correct")

	obj, err = FromString(`{}`)
	a.Nil(err, "err is nil")

	val = obj.SliceOrDefault([]interface{}{true, false, true})
	a.Equal([]interface{}{true, false, true}, val, "val is correct")
}

func Test_MustSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":[true,false,true]}`)
	a.Nil(err, "err is nil")

	val := obj.MustSlice("a")
	a.Equal([]interface{}{true, false, true}, val, "val is correct")
}

func Test_Bool(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`true`)
	a.Nil(err, "err is nil")

	val, err := obj.Bool()
	a.Nil(err, "err is nil")
	a.Equal(true, val, "val is correct")
}

func Test_Bool_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":true}`)
	a.Nil(err, "err is nil")

	val, err := obj.Bool("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Equal(false, val, "val is correct")
}

func Test_Bool_Error(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val, err := obj.Bool()
	a.NotNil(err, "err is not nil")
	a.Equal(false, val, "val is correct")
}

func Test_BoolOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`true`)
	a.Nil(err, "err is nil")

	val := obj.BoolOrDefault(false)
	a.Equal(true, val, "val is correct")

	obj, err = FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val = obj.BoolOrDefault(true)
	a.Equal(true, val, "val is correct")
}

func Test_MustBool(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`true`)
	a.Nil(err, "err is nil")

	val := obj.MustBool()
	a.Equal(true, val, "val is correct")
}

func Test_ID(t *testing.T) {
	a := assert.New(t)

	idGen := NewIDGen()
	id := idGen.MustNew()
	obj := &Json{id}
	a.Equal(id, obj.MustID())
	a.Equal(id, obj.IDOrDefault(id))
	a.Equal(id, (&Json{}).IDOrDefault(id))

	id, err := obj.ID("a")
	a.NotNil(err)

	obj = &Json{true}
	id, err = obj.ID()
	a.Equal(invalidTypeErr, err)

	id = idGen.MustNew()
	obj = &Json{id.String()}
	id, err = obj.ID()
	a.Nil(err)

	id = idGen.MustNew()
	obj = MustFromString(Strf(`{"a": "%s"}`, id.String()))
	id, err = obj.ID("a")
	a.Nil(err)
}

func Test_IDs(t *testing.T) {
	a := assert.New(t)

	idGen := NewIDGen()
	ids := IDs{idGen.MustNew()}
	obj := &Json{ids}
	a.Equal(ids, obj.MustIDs())
	a.Equal(ids, obj.IDsOrDefault(ids))
	a.Nil((&Json{}).IDsOrDefault(ids))
	a.Equal(ids, (&Json{}).IDsOrDefault("a", ids))

	_, err := obj.IDs("a")
	a.NotNil(err)

	a.Equal(ids, (&Json{[]string{ids[0].String()}}).MustIDs())
	a.Equal(ids, (&Json{[]interface{}{ids[0].String()}}).MustIDs())
	a.Equal(ids, (&Json{[]interface{}{ids[0]}}).MustIDs())
	_, err = (&Json{[]interface{}{"not an id"}}).IDs()
	a.NotNil(err)
	_, err = (&Json{[]interface{}{123}}).IDs()
	a.NotNil(err)
	_, err = (&Json{123}).IDs()
	a.NotNil(err)
	_, err = (&Json{[]string{"yolo"}}).IDs()
	a.NotNil(err)
}

func Test_Key(t *testing.T) {
	a := assert.New(t)

	k := MustToKey("yolo")
	obj := &Json{k}
	a.Equal(k, obj.MustKey())
	a.Equal(k, obj.KeyOrDefault(k))
	a.Equal(k, (&Json{}).KeyOrDefault(k))
	a.Equal(k, (&Json{}).KeyOrDefault("yolo"))

	k, err := obj.Key("a")
	a.NotNil(err)

	obj = &Json{true}
	k, err = obj.Key()
	a.Equal(invalidTypeErr, err)

	k = MustToKey("nolo")
	obj = &Json{k.String()}
	k, err = obj.Key()
	a.Nil(err)

	k = MustToKey("bolo")
	obj = MustFromString(Strf(`{"a": "%s"}`, k.String()))
	k, err = obj.Key("a")
	a.Nil(err)
}

func Test_Keys(t *testing.T) {
	a := assert.New(t)

	ks := Keys{"yolo"}
	obj := &Json{ks}
	a.Equal(ks, obj.MustKeys())
	a.Equal(ks, obj.KeysOrDefault(ks))
	a.Nil((&Json{}).KeysOrDefault(ks))
	a.Equal(ks, (&Json{}).KeysOrDefault("a", ks))

	_, err := obj.Keys("a")
	a.NotNil(err)

	a.Equal(ks, (&Json{[]string{ks[0].String()}}).MustKeys())
	a.Equal(ks, (&Json{[]interface{}{ks[0].String()}}).MustKeys())
	a.Equal(ks, (&Json{[]interface{}{ks[0]}}).MustKeys())
	_, err = (&Json{[]interface{}{"not a key"}}).Keys()
	a.NotNil(err)
	_, err = (&Json{[]interface{}{123}}).Keys()
	a.NotNil(err)
	_, err = (&Json{123}).Keys()
	a.NotNil(err)
	_, err = (&Json{[]string{"_yolo_"}}).Keys()
	a.NotNil(err)
}

func Test_String_Error(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`true`)
	a.Nil(err, "err is nil")

	val, err := obj.String()
	a.NotNil(err, "err is not nil")
	a.Equal("", val, "val is correct")
}

func Test_String_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":true}`)
	a.Nil(err, "err is nil")

	val, err := obj.String("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Equal("", val, "val is correct")
}

func Test_MustString_DefaultValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`true`)
	a.Nil(err, "err is nil")

	val := obj.StringOrDefault("hi")
	a.Equal("hi", val, "val is correct")
}

func Test_StringSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`["hi","yo","no"]`)
	a.Nil(err, "err is nil")

	val, err := obj.StringSlice()
	a.Nil(err, "err is nil")
	a.Equal([]string{"hi", "yo", "no"}, val, "val is correct")

	a.Equal([]string(nil), (&Json{}).MustStringSlice())
	a.Equal([]string{}, (&Json{[]string{}}).MustStringSlice())
	_, err = (&Json{}).StringSlice("a")
	a.NotNil(err)
}

func Test_StringSlice_NotSliceError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val, err := obj.StringSlice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_StringSlice_NoneStringValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`["hi",0]`)
	a.Nil(err, "err is nil")

	val, err := obj.StringSlice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_StringSliceOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":["hi"]}`)
	a.Nil(err, "err is nil")

	val := obj.StringSliceOrDefault("a", nil)
	a.Equal([]string{"hi"}, val, "val is correct")

	obj, err = FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val = obj.StringSliceOrDefault([]string{"yo"})
	a.Equal([]string{"yo"}, val, "val is correct")
}

func Test_MustStringSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`["hi"]`)
	a.Nil(err, "err is nil")

	val := obj.MustStringSlice()
	a.Equal([]string{"hi"}, val, "val is correct")
}

func Test_Time(t *testing.T) {
	a := assert.New(t)

	now := Now()
	objStr, err := FromInterface(now).ToString()
	a.Nil(err, "err is nil")

	obj, err := FromString(objStr)
	a.Nil(err, "err is nil")

	val, err := obj.Time()
	a.Nil(err, "err is nil")
	a.Equal(now, val, "val is correct")
}

func Test_Time_Error(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`true`)
	a.Nil(err, "err is nil")

	val, err := obj.Time()
	a.NotNil(err, "err is not nil")
	a.True(val.IsZero(), "val is zero tme")
}

func Test_Time_PathError(t *testing.T) {
	a := assert.New(t)

	obj := FromInterface(map[string]interface{}{"a": Now()})

	val, err := obj.Time("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.True(val.IsZero(), "val is correct")
}

func Test_TimeOrDefault(t *testing.T) {
	a := assert.New(t)

	now := Now()
	obj := FromInterface(now)

	var zero time.Time
	val := obj.TimeOrDefault(zero)
	a.Equal(now, val, "val is correct")

	obj = FromInterface(true)

	now = Now()
	val = obj.TimeOrDefault(now)
	a.Equal(now, val, "val is correct")
}

func Test_MustTime(t *testing.T) {
	a := assert.New(t)

	now := Now()
	obj := FromInterface(now)

	val := obj.MustTime()
	a.Equal(now, val, "val is correct")
}

func Test_TimeSlice(t *testing.T) {
	a := assert.New(t)

	now := Now()
	obj := FromInterface([]time.Time{now, now, now})

	val, err := obj.TimeSlice()
	a.Nil(err, "err is nil")
	a.Equal([]time.Time{now, now, now}, val, "val is correct")

	a.Equal([]time.Time(nil), (&Json{}).MustTimeSlice())
	a.Equal([]time.Time{}, (&Json{[]time.Time{}}).MustTimeSlice())
	nowBytes, _ := now.MarshalText()
	a.Equal([]time.Time{now}, (&Json{[]interface{}{string(nowBytes)}}).MustTimeSlice())
	a.Equal([]time.Time{now}, (&Json{[]string{string(nowBytes)}}).MustTimeSlice())
	_, err = (&Json{[]interface{}{true}}).TimeSlice()
	a.NotNil(err)
	_, err = (&Json{[]string{"abc"}}).TimeSlice()
	a.NotNil(err)
}

func Test_TimeSlice_NotSliceError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val, err := obj.TimeSlice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_TimeSlice_NoneTimeValue(t *testing.T) {
	a := assert.New(t)

	now := Now()
	obj := FromInterface([]interface{}{"hi", now})

	val, err := obj.TimeSlice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_TimeSliceOrDefault(t *testing.T) {
	a := assert.New(t)

	now := Now()
	obj := FromInterface(map[string]interface{}{"a": []time.Time{now}})

	val := obj.TimeSliceOrDefault("a", nil)
	a.Equal([]time.Time{now}, val, "val is correct")

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	def := []time.Time{Now()}
	val = obj.TimeSliceOrDefault("a", def)
	a.Equal(def, val, "val is correct")
}

func Test_MustTimeSlice(t *testing.T) {
	a := assert.New(t)

	now := Now()
	obj := FromInterface(map[string]interface{}{"a": []time.Time{now}})

	val := obj.MustTimeSlice("a")
	a.Equal([]time.Time{now}, val, "val is correct")
}

func Test_Duration(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": "1s"}`)
	a.Nil(err, "err is nil")

	val, err := obj.Duration("a")
	a.Nil(err, "err is nil")
	a.Equal(time.Second, val, "val is correct")
}

func Test_Duration_Error(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": "s"}`)
	a.Nil(err, "err is nil")

	_, err = obj.Duration("a")
	a.NotNil(err, "err is not nil")
}

func Test_Duration_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": "s"}`)
	a.Nil(err, "err is nil")

	_, err = obj.Duration("b")
	a.NotNil(err, "err is not nil")
}

func Test_Duration_FromDuration(t *testing.T) {
	a := assert.New(t)

	dur, err := MustNew().MustSet("a", time.Second).Duration("a")
	a.Nil(err, "err is nil")
	a.Equal(time.Second, dur)
}

func Test_Duration_TypeError(t *testing.T) {
	a := assert.New(t)

	_, err := MustNew().MustSet("a", true).Duration("a")
	a.NotNil(err, "err is not nil")
}

func Test_DurationOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": "1s"}`)
	a.Nil(err, "err is nil")

	val := obj.DurationOrDefault("a", 5*time.Second)
	a.Equal(time.Second, val, "val is correct")

	obj, err = FromString(`{"a": "s"}`)
	a.Nil(err, "err is nil")

	val = obj.DurationOrDefault("a", 5*time.Second)
	a.Equal(5*time.Second, val, "val is correct")
}

func Test_MustDuration(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": "1s"}`)
	a.Nil(err, "err is nil")

	val := obj.MustDuration("a")
	a.Equal(time.Second, val, "val is correct")
}

func Test_DurationSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": ["1s"]}`)
	a.Nil(err, "err is nil")

	val, err := obj.DurationSlice("a")
	a.Nil(err, "err is nil")
	a.Equal(time.Second, val[0], "val is correct")

	a.Equal([]time.Duration(nil), (&Json{nil}).MustDurationSlice())
	a.Equal([]time.Duration{}, (&Json{[]time.Duration{}}).MustDurationSlice())
	_, err = (&Json{nil}).DurationSlice("a")
	a.NotNil(err)
}

func Test_DurationSlice_Error(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": ["s"]}`)
	a.Nil(err, "err is nil")

	_, err = obj.DurationSlice("a")
	a.NotNil(err, "err is not nil")
}

func Test_DurationSlice_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": ["s"]}`)
	a.Nil(err, "err is nil")

	_, err = obj.DurationSlice()
	a.NotNil(err, "err is not nil")
}

func Test_DurationSliceOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": ["1s"]}`)
	a.Nil(err, "err is nil")

	val := obj.DurationSliceOrDefault("a", []time.Duration{5 * time.Second})
	a.Equal(time.Second, val[0], "val is correct")

	obj, err = FromString(`{"a": ["s"]}`)
	a.Nil(err, "err is nil")

	val = obj.DurationSliceOrDefault("a", []time.Duration{5 * time.Second})
	a.Equal([]time.Duration{5 * time.Second}, val, "val is correct")
}

func Test_MustDurationSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a": ["1s"]}`)
	a.Nil(err, "err is nil")

	val := obj.MustDurationSlice("a")
	a.Equal(time.Second, val[0], "val is correct")
}

func Test_Int(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`42`)
	a.Nil(err, "err is nil")

	val, err := obj.Int()
	a.Nil(err, "err is nil")
	a.Equal(42, val, "val is correct")
}

func Test_Int_WithAFloat(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42.3}

	val, err := obj.Int()
	a.Nil(err, "err is nil")
	a.Equal(42, val, "val is correct")
}

func Test_Int_WithAJsonFloat(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`42.3`)
	a.Nil(err, "err is nil")

	val, err := obj.Int()
	a.Nil(err, "err is nil")
	a.Equal(42, val, "val is correct")
}

func Test_Int_WithAnInt(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val, err := obj.Int()
	a.Nil(err, "err is nil")
	a.Equal(42, val, "val is correct")
}

func Test_Int_WithAUint(t *testing.T) {
	a := assert.New(t)

	obj := &Json{uint(42)}

	val, err := obj.Int()
	a.Nil(err, "err is nil")
	a.Equal(42, val, "val is correct")
}

func Test_Int_Error(t *testing.T) {
	a := assert.New(t)

	obj := &Json{"hi"}

	val, err := obj.Int()
	a.NotNil(err, "err is not nil")
	a.Equal(0, val, "val is correct")
}

func Test_IntOrDefault(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.IntOrDefault(24)
	a.Equal(42, val, "val is correct")

	obj = &Json{"hi"}

	val = obj.IntOrDefault(24)
	a.Equal(24, val, "val is correct")
}

func Test_MustInt(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.MustInt()
	a.Equal(42, val, "val is correct")
}

func Test_IntSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val, err := obj.IntSlice()
	a.Nil(err, "err is nil")
	a.Equal([]int{0, 1, 2}, val, "val is correct")
	a.Equal([]int(nil), (&Json{nil}).MustIntSlice())
	a.Equal([]int{}, (&Json{[]int{}}).MustIntSlice())
	_, err = (&Json{nil}).IntSlice("a")
	a.NotNil(err)
}

func Test_IntSlice_NotSliceError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val, err := obj.IntSlice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_IntSlice_NoneIntValue(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,"hi"]`)
	a.Nil(err, "err is nil")

	val, err := obj.IntSlice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_IntSliceOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.IntSliceOrDefault(nil)
	a.Equal([]int{0, 1, 2}, val, "val is correct")

	obj, err = FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val = obj.IntSliceOrDefault([]int{0, 1, 2})
	a.Equal([]int{0, 1, 2}, val, "val is correct")
}

func Test_MustIntSlice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.MustIntSlice()
	a.Equal([]int{0, 1, 2}, val, "val is correct")
}

func Test_Float64_typeError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":true}`)

	_, err = obj.Float64("a")
	a.Equal(invalidTypeErr, err, "error message is correct")
}

func Test_Float64_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":24}`)

	val, err := obj.Float64("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Equal(float64(0), val, "val is correct")
}

func Test_Float64OrDefault(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.Float64OrDefault(float64(24))
	a.Equal(42.0, val, "val is correct")

	obj = &Json{"hi"}

	val = obj.Float64OrDefault(float64(24))
	a.Equal(24.0, val, "val is correct")
}

func Test_MustFloat64(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.MustFloat64()
	a.Equal(42.0, val, "val is correct")
}

func Test_Float64Slice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val, err := obj.Float64Slice()
	a.Nil(err, "err is nil")
	a.Equal([]float64{0.0, 1.0, 2.0}, val, "val is correct")
	a.Equal([]float64(nil), (&Json{nil}).MustFloat64Slice())
	a.Equal([]float64{}, (&Json{[]float64{}}).MustFloat64Slice())
	_, err = (&Json{nil}).Float64Slice("a")
	a.NotNil(err)
}

func Test_Float64Slice_NotSliceError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val, err := obj.Float64Slice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_Float64Slice_NoneFloat64Value(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,"hi"]`)
	a.Nil(err, "err is nil")

	val, err := obj.Float64Slice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_Float64SliceOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.Float64SliceOrDefault(nil)
	a.Equal([]float64{0.0, 1.0, 2.0}, val, "val is correct")

	obj, err = FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val = obj.Float64SliceOrDefault([]float64{0.0, 1.0, 2.0})
	a.Equal([]float64{0.0, 1.0, 2.0}, val, "val is correct")
}

func Test_MustFloat64Slice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.MustFloat64Slice()
	a.Equal([]float64{0.0, 1.0, 2.0}, val, "val is correct")
}

func Test_Int64(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`42`)
	a.Nil(err, "err is nil")

	val, err := obj.Int64()
	a.Nil(err, "err is nil")
	a.Equal(int64(42), val, "val is correct")

	obj, err = FromString(`true`)
	a.Nil(err, "err is nil")

	_, err = obj.Int64()
	a.Equal(invalidTypeErr, err)
}

func Test_Int64_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":42}`)
	a.Nil(err, "err is nil")

	val, err := obj.Int64("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Equal(int64(0), val, "val is correct")
}

func Test_Int64_WithAFloat(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42.3}

	val, err := obj.Int64()
	a.Nil(err, "err is nil")
	a.Equal(int64(42), val, "val is correct")
}

func Test_Int64_WithAJsonFloat(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`42.3`)
	a.Nil(err, "err is nil")

	val, err := obj.Int64()
	a.NotNil(err, "err is not nil")
	a.Equal(int64(0), val, "val is correct")
}

func Test_Int64_WithAnInt64(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val, err := obj.Int64()
	a.Nil(err, "err is nil")
	a.Equal(int64(42), val, "val is correct")
}

func Test_Int64_WithAUint64(t *testing.T) {
	a := assert.New(t)

	obj := &Json{uint64(42)}

	val, err := obj.Int64()
	a.Nil(err, "err is nil")
	a.Equal(int64(42), val, "val is correct")
}

func Test_Int64_Error(t *testing.T) {
	a := assert.New(t)

	obj := &Json{"hi"}

	val, err := obj.Int64()
	a.NotNil(err, "err is not nil")
	a.Equal(int64(0), val, "val is correct")
}

func Test_Int64OrDefault(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.Int64OrDefault(int64(24))
	a.Equal(int64(42), val, "val is correct")

	obj = &Json{"hi"}

	val = obj.Int64OrDefault(int64(24))
	a.Equal(int64(24), val, "val is correct")
}

func Test_MustInt64(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.MustInt64()
	a.Equal(int64(42), val, "val is correct")
}

func Test_Int64Slice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val, err := obj.Int64Slice()
	a.Nil(err, "err is nil")
	a.Equal([]int64{0, 1, 2}, val, "val is correct")
	a.Equal([]int64(nil), (&Json{nil}).MustInt64Slice())
	a.Equal([]int64{}, (&Json{[]int64{}}).MustInt64Slice())
	_, err = (&Json{nil}).Int64Slice("a")
	a.NotNil(err)
}

func Test_Int64Slice_NotSliceError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val, err := obj.Int64Slice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_Int64Slice_NoneInt64Value(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,"hi"]`)
	a.Nil(err, "err is nil")

	val, err := obj.Int64Slice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_Int64SliceOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.Int64SliceDefault(nil)
	a.Equal([]int64{0, 1, 2}, val, "val is correct")

	obj, err = FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val = obj.Int64SliceDefault([]int64{0, 1, 2})
	a.Equal([]int64{0, 1, 2}, val, "val is correct")
}

func Test_MustInt64Slice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.MustInt64Slice()
	a.Equal([]int64{0, 1, 2}, val, "val is correct")
}

func Test_Uint64(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`42`)
	a.Nil(err, "err is nil")

	val, err := obj.Uint64()
	a.Nil(err, "err is nil")
	a.Equal(uint64(42), val, "val is correct")

	obj, err = FromString(`true`)
	a.Nil(err, "err is nil")

	_, err = obj.Uint64()
	a.Equal(invalidTypeErr, err)
}

func Test_Uint64_PathError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`{"a":42}`)
	a.Nil(err, "err is nil")

	val, err := obj.Uint64("a", "b")
	a.NotNil(err, "err is not nil")
	a.Equal([]interface{}{"a"}, ToError(err).Value().(*jsonPathError).FoundPath, "error FoundPath is correct")
	a.Equal([]interface{}{"b"}, ToError(err).Value().(*jsonPathError).MissingPath, "error FoundPath is correct")
	a.Equal("found: [a] missing: [b]", ToError(err).Value().(*jsonPathError).Error(), "error message is correct")
	a.Equal(uint64(0), val, "val is correct")
}

func Test_Uint64_WithAFloat(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42.3}

	val, err := obj.Uint64()
	a.Nil(err, "err is nil")
	a.Equal(uint64(42), val, "val is correct")
}

func Test_Uint64_WithAJsonFloat(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`42.3`)
	a.Nil(err, "err is nil")

	val, err := obj.Uint64()
	a.NotNil(err, "err is not nil")
	a.Equal(uint64(0), val, "val is correct")
}

func Test_Uint64_WithAnUint64(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val, err := obj.Uint64()
	a.Nil(err, "err is nil")
	a.Equal(uint64(42), val, "val is correct")
}

func Test_Uint64_WithAUuint64(t *testing.T) {
	a := assert.New(t)

	obj := &Json{uint64(42)}

	val, err := obj.Uint64()
	a.Nil(err, "err is nil")
	a.Equal(uint64(42), val, "val is correct")
}

func Test_Uint64_Error(t *testing.T) {
	a := assert.New(t)

	obj := &Json{"hi"}

	val, err := obj.Uint64()
	a.NotNil(err, "err is not nil")
	a.Equal(uint64(0), val, "val is correct")
}

func Test_MustUint64OrDefault(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.Uint64OrDefault(uint64(24))
	a.Equal(uint64(42), val, "val is correct")

	obj = &Json{"hi"}

	val = obj.Uint64OrDefault(uint64(24))
	a.Equal(uint64(24), val, "val is correct")
}

func Test_MustUint64(t *testing.T) {
	a := assert.New(t)

	obj := &Json{42}

	val := obj.MustUint64()
	a.Equal(uint64(42), val, "val is correct")
}

func Test_Uint64Slice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val, err := obj.Uint64Slice()
	a.Nil(err, "err is nil")
	a.Equal([]uint64{0, 1, 2}, val, "val is correct")
	a.Equal([]uint64(nil), (&Json{nil}).MustUint64Slice())
	a.Equal([]uint64{}, (&Json{[]uint64{}}).MustUint64Slice())
	_, err = (&Json{nil}).Uint64Slice("a")
	a.NotNil(err)
}

func Test_Uint64Slice_NotSliceError(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val, err := obj.Uint64Slice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_Uint64Slice_NoneUint64Value(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,"hi"]`)
	a.Nil(err, "err is nil")

	val, err := obj.Uint64Slice()
	a.NotNil(err, "err is not nil")
	a.Nil(val, "val is nil")
}

func Test_Uint64SliceOrDefault(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.Uint64SliceOrDefault(nil)
	a.Equal([]uint64{0, 1, 2}, val, "val is correct")

	obj, err = FromString(`"hi"`)
	a.Nil(err, "err is nil")

	val = obj.Uint64SliceOrDefault([]uint64{0, 1, 2})
	a.Equal([]uint64{0, 1, 2}, val, "val is correct")
}

func Test_MustUint64Slice(t *testing.T) {
	a := assert.New(t)

	obj, err := FromString(`[0,1,2]`)
	a.Nil(err, "err is nil")

	val := obj.MustUint64Slice()
	a.Equal([]uint64{0, 1, 2}, val, "val is correct")
}
