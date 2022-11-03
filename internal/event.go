package webscenario

import (
	"context"
	"sync"

	"github.com/yuin/gopher-lua"
)

type EventHandleFunc func(*Tab, *lua.LFunction, *lua.LTable)

type EventHandler struct {
	sync.Mutex

	waiters []chan struct{}
	history []*lua.LTable
	waited  int
	lfunc   *lua.LFunction
	handle  EventHandleFunc
}

func NewEventHandler(handle EventHandleFunc) *EventHandler {
	var h EventHandler
	h.handle = handle
	return &h
}

func (h *EventHandler) Close() error {
	for _, w := range h.waiters {
		close(w)
	}
	return nil
}

func (h *EventHandler) SetFunc(f *lua.LFunction) {
	h.Lock()
	defer h.Unlock()
	h.lfunc = f
}

func (h *EventHandler) IsFuncSet() bool {
	h.Lock()
	defer h.Unlock()
	return h.lfunc != nil
}

func (h *EventHandler) Wait(ctx context.Context) *lua.LTable {
	for {
		h.Lock()

		if h.waited < len(h.history) {
			r := h.history[h.waited]
			h.waited++
			h.Unlock()
			return r
		}

		ch := make(chan struct{})
		h.waiters = append(h.waiters, ch)

		h.Unlock()

		select {
		case <-ch:
			continue
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *EventHandler) Invoke(tab *Tab, event *lua.LTable) {
	h.Lock()
	defer h.Unlock()
	h.history = append(h.history, event)
	h.handle(tab, h.lfunc, event)
	for _, w := range h.waiters {
		close(w)
	}
	h.waiters = nil
}

func (h *EventHandler) Status(L *lua.LState) *lua.LTable {
	h.Lock()
	defer h.Unlock()

	tbl := L.NewTable()
	for _, e := range h.history {
		tbl.Append(e)
	}

	L.SetField(tbl, "_waited", lua.LNumber(h.waited))

	return tbl
}
