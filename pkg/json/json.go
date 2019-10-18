package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
)

var (
	strIsInt         = regexp.MustCompile(`^[1-9][0-9]*$`)
	invalidTypeErr   = errors.New("invalid value type")
	emptyPathPartErr = errors.New("empty path part")
)

func SplitPath(path string) ([]interface{}, error) {
	parts := strings.Split(path, ".")
	finalPath := make([]interface{}, 0, len(parts)+1)
	for _, part := range parts {
		if part == "" {
			return nil, emptyPathPartErr
		}
		if strIsInt.MatchString(part) {
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			finalPath = append(finalPath, num)
		} else {
			finalPath = append(finalPath, part)
		}
	}
	return finalPath, nil
}

func MustSplitPath(path string) []interface{} {
	finalPath, err := SplitPath(path)
	PanicOn(err)
	return finalPath
}

type Json struct {
	data interface{}
}

func New() (*Json, error) {
	return FromString("{}")
}

func MustNew() *Json {
	js, err := New()
	PanicOn(err)
	return js
}

func FromInterface(i interface{}) *Json {
	return &Json{i}
}

func FromString(str string) (*Json, error) {
	return FromBytes([]byte(str))
}

func MustFromString(str string) *Json {
	js, err := FromString(str)
	PanicOn(err)
	return js
}

func FromBytes(b []byte) (*Json, error) {
	return FromReader(bytes.NewReader(b))
}

func MustFromBytes(b []byte) *Json {
	js, err := FromBytes(b)
	PanicOn(err)
	return js
}

func FromFile(file string) (*Json, error) {
	fullPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return FromBytes(data)
}

func MustFromFile(file string) *Json {
	js, err := FromFile(file)
	PanicOn(err)
	return js
}

func FromReader(r io.Reader) (*Json, error) {
	if r == nil {
		return FromString("null")
	}
	rc, ok := r.(io.ReadCloser)
	if !ok {
		rc = ioutil.NopCloser(r)
	}
	return FromReadCloser(rc)
}

func MustFromReader(r io.Reader) *Json {
	js, err := FromReader(r)
	PanicOn(err)
	return js
}

func FromReadCloser(rc io.ReadCloser) (*Json, error) {
	if rc == nil {
		return FromString("null")
	}
	defer rc.Close()
	j := &Json{}
	dec := json.NewDecoder(rc)
	dec.UseNumber()
	err := dec.Decode(&j.data)
	return j, err
}

func MustFromReadCloser(rc io.ReadCloser) *Json {
	js, err := FromReadCloser(rc)
	PanicOn(err)
	return js
}

func (j *Json) ToBytes() ([]byte, error) {
	return j.MarshalJSON()
}

func (j *Json) MustToBytes() []byte {
	bs, err := j.ToBytes()
	PanicOn(err)
	return bs
}

func (j *Json) ToString() (string, error) {
	b, err := j.ToBytes()
	return string(b), err
}

func (j *Json) MustToString() string {
	str, err := j.ToString()
	PanicOn(err)
	return str
}

func (j *Json) ToPrettyBytes() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

func (j *Json) MustToPrettyBytes() []byte {
	bs, err := j.ToPrettyBytes()
	PanicOn(err)
	return bs
}

func (j *Json) ToPrettyString() (string, error) {
	b, err := j.ToPrettyBytes()
	return string(b), err
}

func (j *Json) MustToPrettyString() string {
	str, err := j.ToPrettyString()
	PanicOn(err)
	return str
}

func (j *Json) ToFile(file string, perm os.FileMode) error {
	b, err := j.ToBytes()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, b, perm)
}

func (j *Json) MustToFile(file string, perm os.FileMode) {
	PanicOn(j.ToFile(file, perm))
}

func (j *Json) ToReader() (io.Reader, error) {
	b, err := j.ToBytes()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func (j *Json) MustToReader() io.Reader {
	r, err := j.ToReader()
	PanicOn(err)
	return r
}

func (j *Json) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

func (j *Json) UnmarshalJSON(p []byte) error {
	jNew, err := FromReader(bytes.NewReader(p))
	j.data = jNew.data
	return err
}

func (j *Json) Get(path ...interface{}) (*Json, error) {
	tmp := j
	for i, k := range path {
		if key, ok := k.(string); ok {
			if m, err := tmp.Map(); err == nil {
				if val, ok := m[key]; ok {
					tmp = &Json{val}
				} else {
					return tmp, &jsonPathError{path[:i], path[i:]}
				}
			} else {
				return tmp, &jsonPathError{path[:i], path[i:]}
			}
		} else if index, ok := k.(int); ok {
			if a, err := tmp.Slice(); err == nil {
				if index < 0 || index >= len(a) {
					return tmp, &jsonPathError{path[:i], path[i:]}
				} else {
					tmp = &Json{a[index]}
				}
			} else {
				return tmp, &jsonPathError{path[:i], path[i:]}
			}
		} else {
			return tmp, &jsonPathError{path[:i], path[i:]}
		}
	}
	return tmp, nil
}

func (j *Json) MustGet(path ...interface{}) *Json {
	js, err := j.Get(path...)
	PanicOn(err)
	return js
}

func (j *Json) Set(pathThenValue ...interface{}) error {
	path, val, err := splitPathThenValue(pathThenValue)
	if err != nil {
		return err
	}

	if len(path) == 0 {
		j.data = val
		return nil
	}

	tmp := j

	for i := 0; i < len(path); i++ {
		if key, ok := path[i].(string); ok {
			if m, err := tmp.Map(); err == nil {
				if i == len(path)-1 {
					m[key] = val
				} else {
					_, ok := path[i+1].(string)
					_, exists := m[key]
					if ok && !exists {
						m[key] = map[string]interface{}{}
					}
					tmp = &Json{m[key]}
				}
			} else {
				return &jsonPathError{path[:i], path[i:]}
			}
		} else if index, ok := path[i].(int); ok {
			if a, err := tmp.Slice(); err == nil && index >= 0 && index < len(a) {
				if i == len(path)-1 {
					a[index] = val
				} else {
					tmp = &Json{a[index]}
				}
			} else {
				return &jsonPathError{path[:i], path[i:]}
			}
		} else {
			return &jsonPathError{path[:i], path[i:]}
		}
	}

	return nil
}

func (j *Json) MustSet(pathThenValue ...interface{}) *Json {
	PanicOn(j.Set(pathThenValue...))
	return j
}

func (j *Json) Del(path ...interface{}) error {
	if len(path) == 0 {
		j.data = nil
		return nil
	}

	i := len(path) - 1
	tmp, err := j.Get(path[:i]...)
	if err != nil {
		err.(*jsonPathError).MissingPath = append(err.(*jsonPathError).MissingPath, path[i])
		return err
	}

	if key, ok := path[i].(string); ok {
		if m, err := tmp.Map(); err != nil {
			return &jsonPathError{path[:i], path[i:]}
		} else {
			delete(m, key)
		}
	} else if index, ok := path[i].(int); ok {
		if a, err := tmp.Slice(); err != nil {
			return &jsonPathError{path[:i], path[i:]}
		} else if index < 0 || index >= len(a) {
			return &jsonPathError{path[:i], path[i:]}
		} else {
			a, a[len(a)-1] = append(a[:index], a[index+1:]...), nil
			if i == 0 {
				j.data = a
			} else {
				tmp, _ = j.Get(path[:i-1]...)
				if key, ok := path[i-1].(string); ok {
					tmp.MapOrDefault(nil)[key] = a //is this safe? should be 100% certainty ;)
				} else if index, ok := path[i-1].(int); ok {
					tmp.SliceOrDefault(nil)[index] = a //is this safe? should be 100% certainty ;)
				}
			}
		}
	} else {
		return &jsonPathError{path[:i], path[i:]}
	}
	return nil
}

func (j *Json) MustDel(path ...interface{}) {
	PanicOn(j.Del(path...))
}

func (j *Json) Interface(path ...interface{}) (interface{}, error) {
	tmp, err := j.Get(path...)
	return tmp.data, err
}

func (j *Json) MustInterface(path ...interface{}) interface{} {
	i, err := j.Interface(path...)
	PanicOn(err)
	return i
}

func (j *Json) InterfaceOrDefault(pathThenDefault ...interface{}) interface{} {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.Interface(path...); err == nil {
		return a
	}
	return def
}

func (j *Json) Map(path ...interface{}) (map[string]interface{}, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if tmp.data == nil {
		return map[string]interface{}{}, nil
	}
	if m, ok := tmp.data.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, invalidTypeErr
}

func (j *Json) MustMap(path ...interface{}) map[string]interface{} {
	v, err := j.Map(path...)
	PanicOn(err)
	return v
}

func (j *Json) MapOrDefault(pathThenDefault ...interface{}) map[string]interface{} {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.Map(path...); err == nil {
		return a
	}
	return def.(map[string]interface{})
}

func (j *Json) MapString(path ...interface{}) (map[string]string, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if tmp.data == nil {
		return map[string]string{}, nil
	}
	if m, ok := tmp.data.(map[string]string); ok {
		return m, nil
	}
	if m, ok := tmp.data.(map[string]interface{}); ok {
		ms := map[string]string{}
		for k, v := range m {
			if kStr, ok := v.(string); ok {
				ms[k] = kStr
			} else {
				return nil, invalidTypeErr
			}
		}
		return ms, nil
	}
	return nil, invalidTypeErr
}

func (j *Json) MustMapString(path ...interface{}) map[string]string {
	v, err := j.MapString(path...)
	PanicOn(err)
	return v
}

func (j *Json) MapStringOrDefault(pathThenDefault ...interface{}) map[string]string {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if m, err := j.MapString(path...); err == nil {
		return m
	}
	return def.(map[string]string)
}

func (j *Json) Slice(path ...interface{}) ([]interface{}, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if a, ok := tmp.data.([]interface{}); ok {
		return a, nil
	}
	return nil, invalidTypeErr
}

func (j *Json) MustSlice(path ...interface{}) []interface{} {
	v, err := j.Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) SliceOrDefault(pathThenDefault ...interface{}) []interface{} {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.Slice(path...); err == nil {
		return a
	}
	return def.([]interface{})
}

func (j *Json) Bool(path ...interface{}) (bool, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return false, err
	}
	if s, ok := tmp.data.(bool); ok {
		return s, nil
	}
	return false, invalidTypeErr
}

func (j *Json) MustBool(path ...interface{}) bool {
	v, err := j.Bool(path...)
	PanicOn(err)
	return v
}

func (j *Json) BoolOrDefault(pathThenDefault ...interface{}) bool {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if b, err := j.Bool(path...); err == nil {
		return b
	}
	return def.(bool)
}

func (j *Json) ID(path ...interface{}) (ID, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return ID{}, err
	}
	if i, ok := tmp.data.(ID); ok {
		return i, nil
	}
	if s, ok := tmp.data.(string); ok {
		i := ID{}
		err = i.UnmarshalText([]byte(s))
		return i, err
	}
	return ID{}, invalidTypeErr
}

func (j *Json) MustID(path ...interface{}) ID {
	v, err := j.ID(path...)
	PanicOn(err)
	return v
}

func (j *Json) IDOrDefault(pathThenDefault ...interface{}) ID {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if i, err := j.ID(path...); err == nil {
		return i
	}
	return def.(ID)
}

func (j *Json) String(path ...interface{}) (string, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return "", err
	}
	if s, ok := tmp.data.(string); ok {
		return s, nil
	}
	return "", invalidTypeErr
}

func (j *Json) MustString(path ...interface{}) string {
	v, err := j.String(path...)
	PanicOn(err)
	return v
}

func (j *Json) StringOrDefault(pathThenDefault ...interface{}) string {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if s, err := j.String(path...); err == nil {
		return s
	}
	return def.(string)
}

func (j *Json) StringSlice(path ...interface{}) ([]string, error) {
	js, err := j.Get(path...)
	if err != nil {
		return nil, err
	}

	if js.data == nil {
		return nil, nil
	}
	if data, ok := js.data.([]string); ok {
		return data, nil
	}
	if data, ok := js.data.([]interface{}); ok {
		strs := make([]string, 0, len(data))
		for _, item := range data {
			if str, ok := item.(string); !ok {
				return nil, invalidTypeErr
			} else {
				strs = append(strs, str)
			}
		}
		return strs, nil
	}

	return nil, invalidTypeErr
}

func (j *Json) MustStringSlice(path ...interface{}) []string {
	v, err := j.StringSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) StringSliceOrDefault(pathThenDefault ...interface{}) []string {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.StringSlice(path...); err == nil {
		return a
	}
	return def.([]string)
}

func (j *Json) Time(path ...interface{}) (time.Time, error) {
	var t time.Time
	tmp, err := j.Get(path...)
	if err != nil {
		return t, err
	}
	if t, ok := tmp.data.(time.Time); ok {
		return t, nil
	}
	if tStr, ok := tmp.data.(string); ok {
		err := t.UnmarshalText([]byte(tStr))
		return t, err
	}
	return t, invalidTypeErr
}

func (j *Json) MustTime(path ...interface{}) time.Time {
	v, err := j.Time(path...)
	PanicOn(err)
	return v
}

func (j *Json) TimeOrDefault(pathThenDefault ...interface{}) time.Time {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if t, err := j.Time(path...); err == nil {
		return t
	}
	return def.(time.Time)
}

func (j *Json) TimeSlice(path ...interface{}) ([]time.Time, error) {
	js, err := j.Get(path...)
	if err != nil {
		return nil, err
	}

	if js.data == nil {
		return nil, nil
	}
	if data, ok := js.data.([]time.Time); ok {
		return data, nil
	}
	data, err := js.StringSlice()
	if err != nil {
		return nil, err
	}
	ts := make([]time.Time, 0, len(data))
	for _, item := range data {
		t := time.Time{}
		err := t.UnmarshalText([]byte(item))
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

func (j *Json) MustTimeSlice(path ...interface{}) []time.Time {
	v, err := j.TimeSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) TimeSliceOrDefault(pathThenDefault ...interface{}) []time.Time {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.TimeSlice(path...); err == nil {
		return a
	}
	return def.([]time.Time)
}

func (j *Json) Duration(path ...interface{}) (time.Duration, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return 0, err
	}
	if dur, ok := tmp.data.(time.Duration); ok {
		return dur, nil
	} else if durStr, ok := tmp.data.(string); ok {
		return time.ParseDuration(durStr)
	}
	return 0, errors.New("type assertion/unmarshalling to time.Duration failed")

}

func (j *Json) MustDuration(path ...interface{}) time.Duration {
	v, err := j.Duration(path...)
	PanicOn(err)
	return v
}

func (j *Json) DurationOrDefault(pathThenDefault ...interface{}) time.Duration {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if d, err := j.Duration(path...); err == nil {
		return d
	}
	return def.(time.Duration)
}

func (j *Json) DurationSlice(path ...interface{}) ([]time.Duration, error) {
	js, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if js.data == nil {
		return nil, nil
	}
	if ds, ok := js.data.([]time.Duration); ok {
		return ds, nil
	}
	data, err := js.StringSlice()
	if err != nil {
		return nil, err
	}
	ds := make([]time.Duration, 0, len(data))
	for _, item := range data {
		d, err := time.ParseDuration(item)
		if err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}
	return ds, nil
}

func (j *Json) MustDurationSlice(path ...interface{}) []time.Duration {
	v, err := j.DurationSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) DurationSliceOrDefault(pathThenDefault ...interface{}) []time.Duration {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.DurationSlice(path...); err == nil {
		return a
	}
	return def.([]time.Duration)
}

func (j *Json) Int(path ...interface{}) (int, error) {
	f, err := j.Float64(path...)
	return int(f), err
}

func (j *Json) MustInt(path ...interface{}) int {
	v, err := j.Int(path...)
	PanicOn(err)
	return v
}

func (j *Json) IntOrDefault(pathThenDefault ...interface{}) int {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if i, err := j.Int(path...); err == nil {
		return i
	}
	return def.(int)
}

func (j *Json) IntSlice(path ...interface{}) ([]int, error) {
	js, err := j.Get(path...)
	if err != nil {
		return nil, err
	}

	if js.data == nil {
		return nil, nil
	}
	if data, ok := js.data.([]int); ok {
		return data, nil
	}
	if data, ok := js.data.([]interface{}); ok {
		is := make([]int, 0, len(data))
		for _, item := range data {
			tmp := &Json{item}
			if i, err := tmp.Int(); err != nil {
				return nil, err
			} else {
				is = append(is, i)
			}
		}
		return is, nil
	}

	return nil, invalidTypeErr
}

func (j *Json) MustIntSlice(path ...interface{}) []int {
	v, err := j.IntSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) IntSliceOrDefault(pathThenDefault ...interface{}) []int {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.IntSlice(path...); err == nil {
		return a
	}
	return def.([]int)
}

func (j *Json) Float64(path ...interface{}) (float64, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return 0, err
	}
	switch tmp.data.(type) {
	case string:
		return json.Number(tmp.data.(string)).Float64()
	case json.Number:
		return tmp.data.(json.Number).Float64()
	case float32, float64:
		return reflect.ValueOf(tmp.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(tmp.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(tmp.data).Uint()), nil
	}
	return 0, invalidTypeErr
}

func (j *Json) MustFloat64(path ...interface{}) float64 {
	v, err := j.Float64(path...)
	PanicOn(err)
	return v
}

func (j *Json) Float64OrDefault(pathThenDefault ...interface{}) float64 {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if f, err := j.Float64(path...); err == nil {
		return f
	}
	return def.(float64)
}

func (j *Json) Float64Slice(path ...interface{}) ([]float64, error) {
	js, err := j.Get(path...)
	if err != nil {
		return nil, err
	}

	if js.data == nil {
		return nil, nil
	}
	if data, ok := js.data.([]float64); ok {
		return data, nil
	}
	if data, ok := js.data.([]interface{}); ok {
		fs := make([]float64, 0, len(data))
		for _, item := range data {
			tmp := &Json{item}
			if i, err := tmp.Float64(); err != nil {
				return nil, err
			} else {
				fs = append(fs, i)
			}
		}
		return fs, nil
	}

	return nil, invalidTypeErr
}

func (j *Json) MustFloat64Slice(path ...interface{}) []float64 {
	v, err := j.Float64Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) Float64SliceOrDefault(pathThenDefault ...interface{}) []float64 {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.Float64Slice(path...); err == nil {
		return a
	}
	return def.([]float64)
}

func (j *Json) Int64(path ...interface{}) (int64, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return 0, err
	}
	switch tmp.data.(type) {
	case string:
		return json.Number(tmp.data.(string)).Int64()
	case json.Number:
		return tmp.data.(json.Number).Int64()
	case float32, float64:
		return int64(reflect.ValueOf(tmp.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(tmp.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(tmp.data).Uint()), nil
	}
	return 0, invalidTypeErr
}

func (j *Json) MustInt64(path ...interface{}) int64 {
	v, err := j.Int64(path...)
	PanicOn(err)
	return v
}

func (j *Json) Int64OrDefault(pathThenDefault ...interface{}) int64 {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if i, err := j.Int64(path...); err == nil {
		return i
	}
	return def.(int64)
}

func (j *Json) Int64Slice(path ...interface{}) ([]int64, error) {
	js, err := j.Get(path...)
	if err != nil {
		return nil, err
	}

	if js.data == nil {
		return nil, nil
	}
	if data, ok := js.data.([]int64); ok {
		return data, nil
	}
	if data, ok := js.data.([]interface{}); ok {
		is := make([]int64, 0, len(data))
		for _, item := range data {
			tmp := &Json{item}
			if i, err := tmp.Int64(); err != nil {
				return nil, err
			} else {
				is = append(is, i)
			}
		}
		return is, nil
	}

	return nil, invalidTypeErr
}

func (j *Json) MustInt64Slice(path ...interface{}) []int64 {
	v, err := j.Int64Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) Int64SliceDefault(pathThenDefault ...interface{}) []int64 {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.Int64Slice(path...); err == nil {
		return a
	}
	return def.([]int64)
}

func (j *Json) Uint64(path ...interface{}) (uint64, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return 0, err
	}
	switch tmp.data.(type) {
	case string:
		return strconv.ParseUint(tmp.data.(string), 10, 64)
	case json.Number:
		return strconv.ParseUint(tmp.data.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(tmp.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(tmp.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(tmp.data).Uint(), nil
	}
	return 0, invalidTypeErr
}

func (j *Json) MustUint64(path ...interface{}) uint64 {
	v, err := j.Uint64(path...)
	PanicOn(err)
	return v
}

func (j *Json) Uint64OrDefault(pathThenDefault ...interface{}) uint64 {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if i, err := j.Uint64(path...); err == nil {
		return i
	}
	return def.(uint64)
}

func (j *Json) Uint64Slice(path ...interface{}) ([]uint64, error) {
	js, err := j.Get(path...)
	if err != nil {
		return nil, err
	}

	if js.data == nil {
		return nil, nil
	}
	if data, ok := js.data.([]uint64); ok {
		return data, nil
	}
	if data, ok := js.data.([]interface{}); ok {
		is := make([]uint64, 0, len(data))
		for _, item := range data {
			tmp := &Json{item}
			if i, err := tmp.Uint64(); err != nil {
				return nil, err
			} else {
				is = append(is, i)
			}
		}
		return is, nil
	}

	return nil, invalidTypeErr
}

func (j *Json) MustUint64Slice(path ...interface{}) []uint64 {
	v, err := j.Uint64Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) Uint64SliceOrDefault(pathThenDefault ...interface{}) []uint64 {
	path, def := mustSplitPathThenValue(pathThenDefault)
	if a, err := j.Uint64Slice(path...); err == nil {
		return a
	}
	return def.([]uint64)
}

func splitPathThenValue(pathThenValue []interface{}) ([]interface{}, interface{}, error) {
	if len(pathThenValue) == 0 {
		return nil, nil, fmt.Errorf("no value supplied")
	}
	return pathThenValue[:len(pathThenValue)-1], pathThenValue[len(pathThenValue)-1], nil
}

func mustSplitPathThenValue(pathThenValue []interface{}) ([]interface{}, interface{}) {
	path, val, err := splitPathThenValue(pathThenValue)
	PanicOn(err)
	return path, val
}

type jsonPathError struct {
	FoundPath   []interface{}
	MissingPath []interface{}
}

func (e *jsonPathError) Error() string {
	return fmt.Sprintf("found: %v missing: %v", e.FoundPath, e.MissingPath)
}
