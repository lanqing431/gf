// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Primitive-slice conversion hooks for database records and nested JSON values.

package gdb

import (
	"reflect"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/internal/json"
	"github.com/gogf/gf/v2/util/gconv"
)

// iVal is used for type assert api for Val().
type iVal interface {
	Val() any
}

var (
	// converter is the internal type converter for gdb.
	converter = gconv.NewConverter()
)

func init() {
	converter.RegisterAnyConverterFunc(
		sliceTypeConverterFunc,
		reflect.TypeOf([]string{}),
		reflect.TypeOf([]float32{}),
		reflect.TypeOf([]float64{}),
		reflect.TypeOf([]int{}),
		reflect.TypeOf([]int32{}),
		reflect.TypeOf([]int64{}),
		reflect.TypeOf([]uint{}),
		reflect.TypeOf([]uint32{}),
		reflect.TypeOf([]uint64{}),
	)
}

// GetConverter returns the internal type converter for gdb.
func GetConverter() gconv.Converter {
	return converter
}

// sliceTypeConverterFunc converts wrapped database values and plain JSON values.
func sliceTypeConverterFunc(from any, to reflect.Value) (err error) {
	// Nested JSON values are already unwrapped and must use the same conversion path.
	fromVal := from
	if v, ok := from.(iVal); ok {
		fromVal = v.Val()
	}
	if fromVal == nil {
		return nil
	}
	switch x := fromVal.(type) {
	case []byte:
		dst := to.Addr().Interface()
		err = json.Unmarshal(x, dst)
	case string:
		dst := to.Addr().Interface()
		err = json.Unmarshal([]byte(x), dst)
	default:
		fromType := reflect.TypeOf(fromVal)
		switch fromType.Kind() {
		case reflect.Slice:
			convertOption := gconv.ConvertOption{
				SliceOption:  gconv.SliceOption{ContinueOnError: true},
				MapOption:    gconv.MapOption{ContinueOnError: true},
				StructOption: gconv.StructOption{ContinueOnError: true},
			}
			dv, err := converter.ConvertWithTypeName(fromVal, to.Type().String(), convertOption)
			if err != nil {
				return err
			}
			to.Set(reflect.ValueOf(dv))
		default:
			err = gerror.Newf(
				`unsupported type converting from type "%T" to type "%T"`,
				fromVal, to,
			)
		}
	}
	return err
}
