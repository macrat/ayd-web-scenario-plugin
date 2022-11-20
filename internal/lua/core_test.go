package lua_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func ExampleState_Type() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushString("hello world")

	fmt.Println(L.Type(-1))

	// OUTPUT:
	// string
}

func ExampleState_IsInteger() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushInteger(123)
	fmt.Println(L.IsInteger(-1))

	L.PushNumber(12.3)
	fmt.Println(L.IsInteger(-1))

	// OUTPUT:
	// true
	// false
}

func ExampleState_ToBoolean() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushBoolean(true)
	fmt.Println(L.ToBoolean(-1))

	L.PushInteger(0)
	fmt.Println(L.ToBoolean(-1)) // A number is a true, even if that's 0.

	L.PushNil()
	fmt.Println(L.ToBoolean(-1)) // A nil is a false.

	// OUTPUT:
	// true
	// true
	// false
}

func ExampleState_ToInteger() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushInteger(123)
	fmt.Println(L.ToInteger(-1))

	L.PushNumber(12.3)
	fmt.Println(L.ToInteger(-1)) // ToInteger can't handle float number so it returns 0.

	// OUTPUT:
	// 123
	// 0
}

func ExampleState_ToString() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushString("hello")
	fmt.Printf("%#v\n", L.ToString(-1))

	L.PushNumber(12.3)
	fmt.Printf("%#v\n", L.ToString(-1))

	// OUTPUT:
	// "hello"
	// "12.3"
}

func ExampleState_ToNumber() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushNumber(12.3)
	fmt.Println(L.ToNumber(-1))

	L.PushInteger(123)
	fmt.Println(L.ToNumber(-1)) // ToNumber can handle both of number and integer.

	// OUTPUT:
	// 12.3
	// 123
}

func ExampleState_ToUserdata() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushUserdata([]any{"hello", 123})

	fmt.Println(L.Type(-1))
	fmt.Printf("%#v\n", L.ToUserdata(-1))

	// OUTPUT:
	// userdata
	// []interface {}{"hello", 123}
}

func ExampleState_CreateTable() {
	L, _ := lua.NewState()
	defer L.Close()

	// Make a table on top of the stack.
	L.CreateTable(0, 0)

	// Push a value on top of the stack, and assign into the table by name.
	L.PushNumber(123)
	L.SetField(1, "hello")

	// Push a value on top of the stack, and assign into the table by index.
	L.PushString("fizz")
	L.SetI(1, 3)

	// Get a value by name.
	L.GetField(1, "hello")
	fmt.Printf("table.hello = (%s) %#v\n", L.Type(2), L.ToNumber(2))

	// Get a value by index.
	L.GetI(1, 3)
	fmt.Printf("table[3] = (%s) %#v\n", L.Type(3), L.ToString(3))

	// OUTPUT:
	// table.hello = (number) 123
	// table[3] = (string) "fizz"
}

func ExampleState_Next() {
	L, _ := lua.NewState()
	defer L.Close()

	// Make a table on top of the stack with values.
	L.LoadString("return {2, 4, a=8}")
	L.Call(0, 1)

	// Push the first index, nil.
	L.PushNil()

	// Iterate table using Next method.
	for L.Next(1) {
		// There are two values on the stack, key and value.
		key := L.ToAny(-2)
		value := L.ToAny(-1)

		fmt.Printf("%#v = %#v\n", key, value)

		// Pop the value from the stack. (Don't pop the key)
		L.Pop(1)
	}

	// OUTPUT:
	// 1 = 2
	// 2 = 4
	// "a" = 8
}

func ExampleState_PushFunction() {
	L, _ := lua.NewState()
	defer L.Close()

	// Push a function to the top of the stack.
	L.PushFunction(func(L *lua.State) int {
		L.PushNumber(L.ToNumber(-2) + L.ToNumber(-1))
		return 1
	})

	// Set the function as "f".
	L.SetGlobal("f")

	// Call the function using Lua code.
	L.LoadString("return f(123, 456)")
	L.Call(0, 1)

	// Check result.
	fmt.Println(L.Type(1))
	fmt.Println(L.ToNumber(1))

	// OUTPUT:
	// number
	// 579
}

func ExampleState_GetGlobal() {
	L, _ := lua.NewState()
	defer L.Close()

	// Set global variable.
	L.DoString("x = 123")

	// Get global variable.
	L.GetGlobal("x")

	// There is a value that got, on the top of the stack.
	fmt.Println(L.Type(1))
	fmt.Println(L.ToNumber(1))

	// OUTPUT:
	// number
	// 123
}

func ExampleState_SetGlobal() {
	L, _ := lua.NewState()
	defer L.Close()

	// Push value to set as a global variable.
	L.PushString("hello")

	// Set the top value of the stack as a global variable.
	L.SetGlobal("x")

	// Get the global variable using Lua code.
	L.LoadString("return x")
	L.Call(0, 1)
	fmt.Println(L.Type(1))
	fmt.Println(L.ToString(1))

	// OUTPUT:
	// string
	// hello
}

func ExampleState_SetHook() {
	L, _ := lua.NewState()
	defer L.Close()

	L.SetHook(lua.MaskCall, 0, func(L *lua.State, d lua.Debug) {
		n := d.Name()
		s := d.Source()
		fmt.Printf("%s: %s(%s): %s(%s)\n", d.Event, s.Source, s.What, n.Name, n.What)
	})

	L.DoString(`
		function greeting()
			return "hello world"
		end
		greeting()
	`)

	L.UnsetHook()
	L.DoString(`math.abs(1)`)

	// OUTPUT:
	// call: =<string>(main): ()
	// call: =<string>(Lua): greeting(global)
}

func ExampleState_GetStack() {
	L, _ := lua.NewState()
	defer L.Close()

	L.PushFunction(func(L *lua.State) int {
		d, ok := L.GetStack(0)
		if ok {
			fmt.Println("name:", d.Name().Name)
		}
		return 0
	})
	L.SetGlobal("f")

	L.DoString(`
		f()
	`)

	// OUTPUT:
	// name: f
}

func AssertValue[T any](t *testing.T, L *lua.State, index int, typ lua.Type, value T) {
	t.Helper()

	if tp := L.Type(index); tp != typ {
		t.Fatalf("unexpected type: %s != %s", tp, typ)
	}

	v := L.ToAny(index)
	if !reflect.DeepEqual(v, value) {
		t.Fatalf("unexpected value: (%s) %#v != (%s) %#v", reflect.TypeOf(v), v, reflect.TypeOf(value), value)
	}
}

func NewTestState(t *testing.T) *lua.State {
	t.Helper()

	L, err := lua.NewState()
	if err != nil {
		t.Fatalf("failed to prepare lua state: %s", err)
	}

	t.Cleanup(func() {
		L.Close()
	})

	return L
}

func TestState_pushAndPop(t *testing.T) {
	L := NewTestState(t)

	L.PushBoolean(true)
	AssertValue(t, L, -1, lua.Boolean, true)

	L.PushBoolean(false)
	AssertValue(t, L, -1, lua.Boolean, false)

	L.PushString("hello")
	AssertValue(t, L, -1, lua.String, "hello")

	L.PushNil()
	AssertValue[any](t, L, -1, lua.Nil, nil)

	L.PushNumber(123)
	AssertValue(t, L, -1, lua.Number, float64(123))

	L.PushInteger(123)
	AssertValue(t, L, -1, lua.Number, int64(123))

	if L.GetTop() != 6 {
		t.Fatalf("unexpected number of values in the stack: %d", L.GetTop())
	}

	L.Pop(3)
	if L.GetTop() != 3 {
		t.Fatalf("unexpected number of values in the stack: %d", L.GetTop())
	}
	AssertValue(t, L, -1, lua.String, "hello")

	L.SetTop(1)
	if L.GetTop() != 1 {
		t.Fatalf("unexpected number of values in the stack: %d", L.GetTop())
	}
	AssertValue(t, L, -1, lua.Boolean, true)
}

func TestState_AbsIndex(t *testing.T) {
	L := NewTestState(t)

	assert := func(input, expect int) {
		t.Helper()
		if i := L.AbsIndex(input); i != expect {
			t.Fatalf("expected %d but got %d", expect, i)
		}
	}

	L.PushInteger(1)

	assert(1, 1)
	assert(-1, 1)

	L.PushInteger(2)

	assert(1, 1)
	assert(-1, 2)
	assert(-2, 1)

	L.PushInteger(3)

	assert(1, 1)
	assert(2, 2)
	assert(-1, 3)
	assert(-2, 2)
}

func TestState_readTable(t *testing.T) {
	L := NewTestState(t)

	if err := L.DoString(`x = {2, 4, "8", a=1, nil, 16, b=2}`); err != nil {
		t.Fatalf("failed to do string: %s", err)
	}

	L.GetGlobal("x")
	if L.Type(1) != lua.Table {
		t.Fatalf("unexpected type of value on the top of stack: %s", L.Type(1))
	}

	L.PushNil()
	if L.Type(2) != lua.Nil {
		t.Fatalf("unexpected type of value on the top of stack: %s", L.Type(2))
	}

	type KV struct {
		Key   any
		Value any
	}
	var nexts []KV
	for i := 0; L.Next(1); i++ {
		nexts = append(nexts, KV{
			Key:   L.ToAny(-2),
			Value: L.ToAny(-1),
		})
		L.Pop(1)
	}
	nextExpects := []KV{
		{int64(1), int64(2)},
		{int64(2), int64(4)},
		{int64(3), "8"},
		{int64(5), int64(16)},
	}
	nextExpectsA := append(nextExpects, KV{"a", int64(1)}, KV{"b", int64(2)})
	nextExpectsB := append(nextExpects, KV{"b", int64(2)}, KV{"a", int64(1)})
	if !reflect.DeepEqual(nexts, nextExpectsA) && !reflect.DeepEqual(nexts, nextExpectsB) {
		t.Fatalf("next results was different\nexpected: %#v\n but got: %#v", nextExpects, nexts)
	}

	if l := L.Len(1); l != 5 {
		t.Fatalf("unexpected length: %d", l)
	}

	expects := []any{int64(2), int64(4), "8", nil, int64(16)}
	types := []lua.Type{lua.Number, lua.Number, lua.String, lua.Nil, lua.Number}
	for i := 0; i < 5; i++ {
		if typ := L.GetI(1, i+1); typ != types[i] {
			t.Fatalf("%d: unexpected type: %s != %s", i, types[i], typ)
		}
		AssertValue(t, L, 2, types[i], expects[i])
		L.SetTop(1)
	}
}

func TestState_PushFunction(t *testing.T) {
	L := NewTestState(t)

	L.PushFunction(func(L *lua.State) int {
		if n := L.GetTop(); n != 2 {
			t.Fatalf("unexpected stack length: %d", n)
		}
		L.PushInteger(L.ToInteger(1) + L.ToInteger(2))
		if n := L.GetTop(); n != 3 {
			t.Fatalf("unexpected stack length: %d", n)
		}
		return 1
	})

	L.PushInteger(1)
	L.PushInteger(2)

	if n := L.GetTop(); n != 3 {
		t.Fatalf("unexpected stack length: %d", n)
	}

	if err := L.Call(2, 1); err != nil {
		t.Fatalf("failed to call function: %s", err)
	}

	if n := L.GetTop(); n != 1 {
		t.Fatalf("unexpected stack length: %d", n)
	}

	if typ := L.Type(1); typ != lua.Number {
		t.Fatalf("unexpected type: %s", typ)
	}

	if !L.IsInteger(1) {
		t.Fatalf("the result should be a integer but not")
	}

	if i := L.ToInteger(1); i != 3 {
		t.Fatalf("unexpected result: %d", i)
	}
}

func TestState_GetMetatable(t *testing.T) {
	L := NewTestState(t)

	if err := L.DoString(`
		x = {}
		setmetatable(x, {foo="bar"})
	`); err != nil {
		t.Fatalf("failed to do string: %s", err)
	}

	L.GetGlobal("x")

	if !L.GetMetatable(1) {
		t.Fatalf("failed to get metatable")
	}

	if typ := L.GetField(2, "foo"); typ != lua.String {
		t.Fatalf("unexpected value in metatable found: %s", typ)
	}

	if s := L.ToString(3); s != "bar" {
		t.Fatalf("unexpected string: %s", s)
	}
}

func TestState_SetMetatable(t *testing.T) {
	L := NewTestState(t)

	L.NewTypeMetatable("setmeta_test")
	L.PushFunction(func(L *lua.State) int {
		if typ := L.Type(1); typ != lua.Table {
			t.Fatalf("unexpected type: %s", typ)
		}
		L.GetField(1, "i")

		if typ := L.Type(2); typ != lua.Number {
			t.Fatalf("unexpected field type: %s", typ)
		}

		L.PushInteger(L.ToInteger(2) + 1)
		L.SetField(1, "i")

		L.PushInteger(L.ToInteger(2) + 1)
		return 1
	})
	L.SetField(1, "__call")
	L.SetTop(0)

	L.CreateTable(0, 0)
	L.PushInteger(0)
	L.SetField(1, "i")

	L.GetTypeMetatable("setmeta_test")
	L.SetMetatable(1)

	L.SetGlobal("x")

	for i := int64(1); i <= 5; i++ {
		L.GetGlobal("x")
		if err := L.Call(0, 1); err != nil {
			t.Fatalf("failed to call: %s", err)
		}
		if typ := L.Type(1); typ != lua.Number {
			t.Fatalf("unexpected return type: %s", typ)
		}
		if n := L.ToInteger(1); n != i {
			t.Fatalf("unexpected number: %d != %d", n, i)
		}
		L.Pop(1)
	}
}

func TestState_Error(t *testing.T) {
	L := NewTestState(t)

	throwed := errors.New("test error")

	L.PushFunction(func(L *lua.State) int {
		L.Error(1, throwed)
		return 0
	})
	L.SetGlobal("f")

	err := L.DoString(`
		function g()
			f()
		end
		g()
	`)
	e, ok := err.(lua.Error)
	if !ok {
		t.Fatalf("unexpected error found: %s", err)
	}

	if !errors.Is(e.Err, throwed) {
		t.Errorf("unexpected error: %s", e)
	}

	if e.ChunkName != "<string>" {
		t.Errorf("unexpected chunk name: %q", e.ChunkName)
	}

	if e.CurrentLine != 3 {
		t.Errorf("unexpected current line: %d", e.CurrentLine)
	}

	trace := strings.Join([]string{
		`stack traceback:`,
		`	<string>:3: in function 'g'`,
		`	<string>:5: in main chunk`,
	}, "\n")
	if e.Traceback != trace {
		t.Errorf("unexpected traceback:\n%s", e.Traceback)
	}

	msg := "<string>:3: test error\n" + trace
	if e.Error() != msg {
		t.Errorf("unexpected message:\n%s", e.Error())
	}
}

func TestState_Error_rethrow(t *testing.T) {
	L := NewTestState(t)

	throwed := errors.New("test error")

	L.PushFunction(func(L *lua.State) int {
		L.Error(1, L.WrapError(1, throwed))
		return 0
	})
	L.SetGlobal("f")

	err := L.DoString(`
		f()
	`)
	e, ok := err.(lua.Error)
	if !ok {
		t.Fatalf("unexpected error found: %s", err)
	}

	if !errors.Is(e.Err, throwed) {
		t.Errorf("unexpected error: %s", e)
	}

	msg := strings.Join([]string{
		`<string>:2: test error`,
		`stack traceback:`,
		`	<string>:2: in main chunk`,
	}, "\n")
	if e.Error() != msg {
		t.Errorf("unexpected message:\n%s", e.Error())
	}
}

func TestState_Where(t *testing.T) {
	L := NewTestState(t)

	called := false
	L.PushFunction(func(L *lua.State) int {
		if n := L.GetTop(); n != 1 {
			t.Fatalf("unexpected number of values on the stack: %d", n)
		}

		c, l := L.Where(0)
		if c != "" {
			t.Fatalf("unexpected chunk name: %q", c)
		}
		if l != 0 {
			t.Fatalf("unexpected current line: %d", l)
		}

		c, l = L.Where(1)
		if c != "<string>" {
			t.Fatalf("unexpected chunk name: %q", c)
		}
		if l != 2 {
			t.Fatalf("unexpected current line: %d", l)
		}

		if n := L.GetTop(); n != 1 {
			t.Fatalf("unexpected number of values on the stack: %d", n)
		}
		if L.ToString(1) != "a" {
			t.Fatalf("stack seems broken: %#v", L.ToAny(1))
		}

		called = true

		return 0
	})
	L.SetGlobal("f")
	err := L.DoString(`
		f('a')
	`)
	if err != nil {
		t.Fatalf("failed to call: %s", err)
	}

	if !called {
		t.Fatalf("function has not called")
	}
}
