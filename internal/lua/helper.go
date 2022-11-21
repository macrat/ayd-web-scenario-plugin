package lua

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func (L *State) Pop(n int) {
	L.SetTop(-n - 1)
}

func (L *State) Remove(index int) {
	L.Rotate(index, -1)
	L.Pop(1)
}

func (L *State) Replace(index int) {
	L.Copy(-1, index)
	L.Pop(1)
}

func (L *State) Swap(a, b int) {
	a, b = L.AbsIndex(a), L.AbsIndex(b)
	L.PushValue(a)
	L.Copy(b, a)
	L.Replace(b)
}

func (L *State) toTable(index int, toAny func(index int) any) any {
	index = L.AbsIndex(index)

	values := make(map[string]any)
	array := make([]any, L.Len(-1))
	intOnly := true

	L.PushNil()
	for L.Next(index) {
		x := toAny(-1)

		values[L.ToString(-2)] = x
		if intOnly && L.IsInteger(-2) && int(L.ToInteger(-2)) <= len(array) {
			array[L.ToInteger(-2)-1] = x
		} else {
			intOnly = false
		}
		L.Pop(1)
	}

	if !intOnly || L.Len(-1) < int64(len(values)) {
		return values
	}

	return array
}

type IncompatibleValue struct {
	S string
}

func (v IncompatibleValue) String() string {
	return v.S
}

func (v IncompatibleValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.S)
}

func (L *State) ToAny(index int) any {
	switch L.Type(index) {
	case Nil, None:
		return nil
	case Boolean:
		return L.ToBoolean(index)
	case Number:
		if L.IsInteger(index) {
			return L.ToInteger(index)
		} else {
			return L.ToNumber(index)
		}
	case String:
		return L.ToString(index)
	case Table:
		return L.toTable(index, L.ToAny)
	case Userdata:
		return L.ToUserdata(index)
	default:
		return IncompatibleValue{L.ToString(index)}
	}
}

func (L *State) ToAnyButInteger(index int) any {
	switch L.Type(index) {
	case Nil, None:
		return nil
	case Boolean:
		return L.ToBoolean(index)
	case Number:
		return L.ToNumber(index)
	case String:
		return L.ToString(index)
	case Table:
		return L.toTable(index, L.ToAnyButInteger)
	case Userdata:
		return L.ToUserdata(index)
	default:
		return IncompatibleValue{L.ToString(index)}
	}
}

func (L *State) pushArray(value reflect.Value) {
	if value.IsNil() {
		L.PushNil()
	} else {
		L.CreateTable(value.Len(), 0)
		for i := 0; i < value.Len(); i++ {
			L.PushAny(value.Index(i).Interface())
			L.SetI(-2, i+1)
		}
	}
}

func (L *State) pushMap(value reflect.Value) {
	if value.IsNil() {
		L.PushNil()
	} else {
		L.CreateTable(0, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			L.PushAny(iter.Value().Interface())
			L.SetField(-2, fmt.Sprint(iter.Key().Interface()))
		}
	}
}

func (L *State) PushAny(value any) {
	if value == nil {
		L.PushNil()
		return
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Bool:
		L.PushBoolean(v.Bool())
	case reflect.Int, reflect.Int32, reflect.Int64:
		L.PushInteger(v.Int())
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		L.PushInteger(int64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		L.PushNumber(v.Float())
	case reflect.String:
		L.PushString(v.String())
	case reflect.Array, reflect.Slice:
		L.pushArray(v)
	case reflect.Map:
		L.pushMap(v)
	case reflect.Func:
		if f, ok := value.(func(*State) int); ok {
			L.PushFunction(f)
		} else {
			L.PushUserdata(value)
		}
	default:
		L.PushUserdata(v.Interface())
	}
}

func (L *State) AssertType(index int, typ Type) {
	if actual := L.Type(index); actual != typ {
		L.ArgErrorf(index, "%s expected, got %s", typ, actual)
	}
}

func (L *State) CheckBoolean(index int) bool {
	L.AssertType(index, Boolean)
	return L.ToBoolean(index)
}

func (L *State) CheckInteger(index int) int64 {
	if !L.IsInteger(index) {
		L.ArgErrorf(index, "integer expected, got %s", L.Type(index))
	}
	return L.ToInteger(index)
}

func (L *State) CheckString(index int) string {
	L.AssertType(index, String)
	return L.ToString(index)
}

func (L *State) CheckNumber(index int) float64 {
	L.AssertType(index, Number)
	return L.ToNumber(index)
}

func (L *State) CheckUserdata(index int) any {
	L.AssertType(index, Userdata)
	return L.ToUserdata(index)
}

func (L *State) SetBoolean(index int, name string, value bool) {
	index = L.AbsIndex(index)
	L.PushBoolean(value)
	L.SetField(index, name)
}

func (L *State) SetInteger(index int, name string, value int64) {
	index = L.AbsIndex(index)
	L.PushInteger(value)
	L.SetField(index, name)
}

func (L *State) SetString(index int, name string, value string) {
	index = L.AbsIndex(index)
	L.PushString(value)
	L.SetField(index, name)
}

func (L *State) SetNil(index int, name string) {
	index = L.AbsIndex(index)
	L.PushNil()
	L.SetField(index, name)
}

func (L *State) SetNumber(index int, name string, value float64) {
	index = L.AbsIndex(index)
	L.PushNumber(value)
	L.SetField(index, name)
}

func (L *State) SetFunction(index int, name string, f GFunction) {
	index = L.AbsIndex(index)
	L.PushFunction(f)
	L.SetField(index, name)
}

func (L *State) SetFuncs(index int, funcs map[string]GFunction) {
	index = L.AbsIndex(index)

	for name, f := range funcs {
		L.PushFunction(f)
		L.SetField(index, name)
	}
}

func (L *State) Do(r io.Reader, name string, isFile bool) error {
	err := L.Load(r, name, isFile)
	if err != nil {
		return err
	}
	return L.Call(0, 0)
}

func (L *State) LoadString(code string) error {
	return L.Load(strings.NewReader(code), "<string>", false)
}

func (L *State) DoString(code string) error {
	return L.Do(strings.NewReader(code), "<string>", false)
}

// CallWithContext calls Lua function with context.
// Don't use this function with SetHook because it uses SetHook internally.
func (L *State) CallWithContext(ctx context.Context, nargs, nret int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	L.SetHook(MaskLine|MaskCount, 1000, func(L *State, d Debug) {
		select {
		case <-ctx.Done():
			panic(ctx.Err())
		default:
		}
	})

	return L.Call(nargs, nret)
}

// DoWithContext is the same as Do but uses context.
// Please see also CallWithContext.
func (L *State) DoWithContext(ctx context.Context, r io.Reader, name string, isFile bool) error {
	err := L.Load(r, name, isFile)
	if err != nil {
		return err
	}
	return L.CallWithContext(ctx, 0, 0)
}

// Errorf raises an error as same as Error, but it accepts format string like fmt.Sprintf.
func (L *State) Errorf(level int, format string, args ...any) {
	L.Error(level, fmt.Errorf(format, args...))
}

func (L *State) ArgErrorf(arg int, format string, args ...any) {
	if d, ok := L.GetStack(0); ok {
		funcName := d.Name().Name
		L.Errorf(1, "bad argument #%d to '%s' (%s)", arg, funcName, fmt.Sprintf(format, args...))
	} else {
		L.Errorf(1, "bad argument #%d (%s)", arg, fmt.Sprintf(format, args...))
	}
}
