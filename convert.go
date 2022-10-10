package main

import (
	"fmt"
	"math"
	"reflect"

	"github.com/yuin/gopher-lua"
)

func asArray(t *lua.LTable) ([]any, bool) {
	isArray := true
	values := make(map[int]lua.LValue)
	t.ForEach(func(k, v lua.LValue) {
		if n, ok := k.(lua.LNumber); ok {
			if math.Mod(float64(n), 1) != 0 {
				isArray = false
			} else {
				values[int(n)] = v
			}
		} else {
			isArray = false
		}
	})
	if !isArray {
		return nil, false
	}
	result := make([]any, len(values))
	for i := 1; i <= len(values); i++ {
		v, ok := values[i]
		if !ok {
			return nil, false
		}
		result[i-1] = UnpackLValue(v)
	}
	return result, true
}

func UnpackLValue(v lua.LValue) any {
	switch x := v.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return x == lua.LTrue
	case lua.LNumber:
		return float64(x)
	case lua.LString:
		return string(x)
	case *lua.LTable:
		if array, ok := asArray(x); ok {
			return array
		}

		values := make(map[string]any)
		x.ForEach(func(k, v lua.LValue) {
			values[k.String()] = UnpackLValue(v)
		})
		return values
	default:
		return x.String()
	}
}

func PackLValue(L *lua.LState, value any) lua.LValue {
	if value == nil {
		return lua.LNil
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Bool:
		return lua.LBool(v.Bool())
	case reflect.Int, reflect.Int32, reflect.Int64:
		return lua.LNumber(float64(v.Int()))
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(float64(v.Float()))
	case reflect.String:
		return lua.LString(v.String())
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			return lua.LNil
		}

		xs := L.NewTable()
		for i := 0; i < v.Len(); i++ {
			xs.Append(PackLValue(L, v.Index(i).Interface()))
		}
		return xs
	case reflect.Map:
		if v.IsNil() {
			return lua.LNil
		}

		xs := L.NewTable()
		iter := v.MapRange()
		for iter.Next() {
			L.SetField(xs, fmt.Sprint(iter.Key().Interface()), PackLValue(L, iter.Value().Interface()))
		}
		return xs
	default:
		ud := L.NewUserData()
		ud.Value = v.Interface()
		return ud
	}
}
