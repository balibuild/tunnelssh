package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func stringCat(sv []string) string {
	var l int
	for _, s := range sv {
		l += len(s)
	}
	var buf strings.Builder
	buf.Grow(l)
	for _, s := range sv {
		_, _ = buf.WriteString(s)
	}
	return buf.String()
}

// StrCat Merges given strings returning the merged result as a string.
func StrCat(s ...string) string {
	return stringCat(s)
}

// ErrorCat cat error
func ErrorCat(s ...string) error {
	return errors.New(stringCat(s))
}

// StrCatEx todo
func StrCatEx(iv ...interface{}) string {
	if len(iv) == 0 {
		return ""
	}
	sv := make([]string, 0, len(iv))
	for _, i := range iv {
		switch f := i.(type) {
		case bool:
			if f {
				sv = append(sv, "true")
			} else {
				sv = append(sv, "false")
			}
		case float32:
			sv = append(sv, strconv.FormatFloat(float64(f), 'g', -1, 10))
		case float64:
			sv = append(sv, strconv.FormatFloat(f, 'g', -1, 10))
		case string:
			sv = append(sv, f)
		case int16:
			sv = append(sv, strconv.FormatInt(int64(f), 10))
		case int32:
			sv = append(sv, strconv.FormatInt(int64(f), 10))
		case int:
			sv = append(sv, strconv.FormatInt(int64(f), 10))
		case int64:
			sv = append(sv, strconv.FormatInt(int64(f), 10))
		case uint16:
			sv = append(sv, strconv.FormatUint(uint64(f), 10))
		case uint32:
			sv = append(sv, strconv.FormatUint(uint64(f), 10))
		case uint:
			sv = append(sv, strconv.FormatUint(uint64(f), 10))
		case uint64:
			sv = append(sv, strconv.FormatUint(uint64(f), 10))
		default:
			sv = append(sv, fmt.Sprintf("%v", i))
		}
	}
	return stringCat(sv)
}
