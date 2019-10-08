package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/0xor1/wtf/pkg/core"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	if len(pathThenValue) == 0 {
		return fmt.Errorf("no value supplied")
	}
	path := pathThenValue[:len(pathThenValue)-1]
	val := pathThenValue[len(pathThenValue)-1]
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

func (j *Json) InterfaceOrDefault(def interface{}, path ...interface{}) interface{} {
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
	if m, ok := tmp.data.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

func (j *Json) MustMap(path ...interface{}) map[string]interface{} {
	v, err := j.Map(path...)
	PanicOn(err)
	return v
}

func (j *Json) MapOrDefault(def map[string]interface{}, path ...interface{}) map[string]interface{} {
	if a, err := j.Map(path...); err == nil {
		return a
	}
	return def
}

func (j *Json) MapString(path ...interface{}) (map[string]string, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if m, ok := tmp.data.(map[string]interface{}); ok {
		ms := map[string]string{}
		for k, v := range m {
			if kStr, ok := v.(string); ok {
				ms[k] = kStr
			} else {
				return nil, errors.New("type assertion of map value to string failed")
			}
		}
		return ms, nil
	}
	return nil, errors.New("type assertion to map[string]string{} failed")
}

func (j *Json) MustMapString(path ...interface{}) map[string]string {
	v, err := j.MapString(path...)
	PanicOn(err)
	return v
}

func (j *Json) MapStringOrDefault(def map[string]string, path ...interface{}) map[string]string {
	if m, err := j.MapString(path...); err == nil {
		return m
	}
	return def
}

func (j *Json) Slice(path ...interface{}) ([]interface{}, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if a, ok := tmp.data.([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

func (j *Json) MustSlice(path ...interface{}) []interface{} {
	v, err := j.Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) SliceOrDefault(def []interface{}, path ...interface{}) []interface{} {
	if a, err := j.Slice(path...); err == nil {
		return a
	}
	return def
}

func (j *Json) Bool(path ...interface{}) (bool, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return false, err
	}
	if s, ok := tmp.data.(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

func (j *Json) MustBool(path ...interface{}) bool {
	v, err := j.Bool(path...)
	PanicOn(err)
	return v
}

func (j *Json) BoolOrDefault(def bool, path ...interface{}) bool {
	if b, err := j.Bool(path...); err == nil {
		return b
	}
	return def
}

func (j *Json) String(path ...interface{}) (string, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return "", err
	}
	if s, ok := tmp.data.(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

func (j *Json) MustString(path ...interface{}) string {
	v, err := j.String(path...)
	PanicOn(err)
	return v
}

func (j *Json) StringOrDefault(def string, path ...interface{}) string {
	if s, err := j.String(path...); err == nil {
		return s
	}
	return def
}

func (j *Json) StringSlice(path ...interface{}) ([]string, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]string, 0, len(arr))
	for _, a := range arr {
		if s, ok := a.(string); a == nil || !ok {
			return nil, errors.New("none string value encountered")
		} else {
			retArr = append(retArr, s)
		}
	}
	return retArr, nil
}

func (j *Json) MustStringSlice(path ...interface{}) []string {
	v, err := j.StringSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) StringSliceOrDefault(def []string, path ...interface{}) []string {
	if a, err := j.StringSlice(path...); err == nil {
		return a
	}
	return def
}

func (j *Json) Time(path ...interface{}) (time.Time, error) {
	var t time.Time
	tmp, err := j.Get(path...)
	if err != nil {
		return t, err
	}
	if t, ok := tmp.data.(time.Time); ok {
		return t, nil
	} else if tStr, ok := tmp.data.(string); ok {
		if t.UnmarshalText([]byte(tStr)) == nil {
			return t, nil
		}
	}
	return t, errors.New("type assertion/unmarshalling to time.Time failed")
}

func (j *Json) MustTime(path ...interface{}) time.Time {
	v, err := j.Time(path...)
	PanicOn(err)
	return v
}

func (j *Json) TimeOrDefault(def time.Time, path ...interface{}) time.Time {
	if t, err := j.Time(path...); err == nil {
		return t
	}
	return def
}

func (j *Json) TimeSlice(path ...interface{}) ([]time.Time, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]time.Time, 0, len(arr))
	for _, a := range arr {
		if s, ok := a.(time.Time); a == nil || !ok {
			return nil, errors.New("none time.Time value encountered")
		} else {
			retArr = append(retArr, s)
		}
	}
	return retArr, nil
}

func (j *Json) MustTimeSlice(path ...interface{}) []time.Time {
	v, err := j.TimeSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) TimeSliceOrDefault(def []time.Time, path ...interface{}) []time.Time {
	if a, err := j.TimeSlice(path...); err == nil {
		return a
	}
	return def
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

func (j *Json) DurationOrDefault(def time.Duration, path ...interface{}) time.Duration {
	if d, err := j.Duration(path...); err == nil {
		return d
	}
	return def
}

func (j *Json) DurationSlice(path ...interface{}) ([]time.Duration, error) {
	arr, err := j.StringSlice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]time.Duration, 0, len(arr))
	for _, a := range arr {
		if d, err := time.ParseDuration(a); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, d)
		}
	}
	return retArr, nil
}

func (j *Json) MustDurationSlice(path ...interface{}) []time.Duration {
	v, err := j.DurationSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) DurationSliceOrDefault(def []time.Duration, path ...interface{}) []time.Duration {
	if a, err := j.DurationSlice(path...); err == nil {
		return a
	}
	return def
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

func (j *Json) IntOrDefault(def int, path ...interface{}) int {
	if i, err := j.Int(path...); err == nil {
		return i
	}
	return def
}

func (j *Json) IntSlice(path ...interface{}) ([]int, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]int, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if i, err := tmp.Int(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, i)
		}
	}
	return retArr, nil
}

func (j *Json) MustIntSlice(path ...interface{}) []int {
	v, err := j.IntSlice(path...)
	PanicOn(err)
	return v
}

func (j *Json) IntSliceOrDefault(def []int, path ...interface{}) []int {
	if a, err := j.IntSlice(path...); err == nil {
		return a
	}
	return def
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

func (j *Json) Float64OrDefault(def float64, path ...interface{}) float64 {
	if f, err := j.Float64(path...); err == nil {
		return f
	}
	return def
}

func (j *Json) Float64Slice(path ...interface{}) ([]float64, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]float64, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if f, err := tmp.Float64(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, f)
		}
	}
	return retArr, nil
}

func (j *Json) MustFloat64Slice(path ...interface{}) []float64 {
	v, err := j.Float64Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) Float64SliceOrDefault(def []float64, path ...interface{}) []float64 {
	if a, err := j.Float64Slice(path...); err == nil {
		return a
	}
	return def
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

func (j *Json) Int64OrDefault(def int64, path ...interface{}) int64 {
	if i, err := j.Int64(path...); err == nil {
		return i
	}
	return def
}

func (j *Json) Int64Slice(path ...interface{}) ([]int64, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]int64, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if i, err := tmp.Int64(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, i)
		}
	}
	return retArr, nil
}

func (j *Json) MustInt64Slice(path ...interface{}) []int64 {
	v, err := j.Int64Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) Int64SliceDefault(def []int64, path ...interface{}) []int64 {
	if a, err := j.Int64Slice(path...); err == nil {
		return a
	}
	return def
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

func (j *Json) Uint64OrDefault(def uint64, path ...interface{}) uint64 {
	if i, err := j.Uint64(path...); err == nil {
		return i
	}
	return def
}

func (j *Json) Uint64Slice(path ...interface{}) ([]uint64, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]uint64, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if u, err := tmp.Uint64(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, u)
		}
	}
	return retArr, nil
}

func (j *Json) MustUint64Slice(path ...interface{}) []uint64 {
	v, err := j.Uint64Slice(path...)
	PanicOn(err)
	return v
}

func (j *Json) Uint64SliceOrDefault(def []uint64, path ...interface{}) []uint64 {
	if a, err := j.Uint64Slice(path...); err == nil {
		return a
	}
	return def
}

type jsonPathError struct {
	FoundPath   []interface{}
	MissingPath []interface{}
}

func (e *jsonPathError) Error() string {
	return fmt.Sprintf("found: %v missing: %v", e.FoundPath, e.MissingPath)
}
