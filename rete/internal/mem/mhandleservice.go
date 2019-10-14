package mem

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/project-flogo/rules/common/model"
	"github.com/project-flogo/rules/rete/common"
	"github.com/project-flogo/rules/rete/internal/types"
)

type handleServiceImpl struct {
	types.NwServiceImpl
	allHandles map[string]*reteHandleImpl
	rand.Source
	sync.RWMutex
}

func NewHandleCollection(nw types.Network, config common.Config) types.HandleService {
	hc := handleServiceImpl{
		NwServiceImpl: types.NwServiceImpl{
			Nw: nw,
		},
		allHandles: make(map[string]*reteHandleImpl),
		Source:     rand.NewSource(time.Now().UnixNano()),
	}
	return &hc
}

func (hc *handleServiceImpl) Init() {
}

func (hc *handleServiceImpl) RemoveHandle(tuple model.Tuple) types.ReteHandle {
	hc.Lock()
	defer hc.Unlock()
	rh, found := hc.allHandles[tuple.GetKey().String()]
	if found {
		delete(hc.allHandles, tuple.GetKey().String())
		return rh
	}
	return nil
}

func (hc *handleServiceImpl) GetHandle(ctx context.Context, tuple model.Tuple) types.ReteHandle {
	hc.RLock()
	defer hc.RUnlock()
	return hc.allHandles[tuple.GetKey().String()]
}

func (hc *handleServiceImpl) GetHandleByKey(ctx context.Context, key model.TupleKey) types.ReteHandle {
	hc.RLock()
	defer hc.RUnlock()
	return hc.allHandles[key.String()]
}

func (hc *handleServiceImpl) GetOrCreateLockedHandle(nw types.Network, tuple model.Tuple) (types.ReteHandle, bool) {
	hc.Lock()
	defer hc.Unlock()
	id := hc.Int63()
	h, found := hc.allHandles[tuple.GetKey().String()]
	if !found {
		h = newReteHandleImpl(nw, tuple, types.ReteHandleStatusCreating, id)
		hc.allHandles[tuple.GetKey().String()] = h
		return h, false
	}

	if atomic.CompareAndSwapInt64(&h.id, -1, id) {
		return h, false
	}

	return nil, true
}

func (hc *handleServiceImpl) GetLockedHandle(nw types.Network, tuple model.Tuple) (types.ReteHandle, bool) {
	hc.Lock()
	defer hc.Unlock()
	id := hc.Int63()
	h, found := hc.allHandles[tuple.GetKey().String()]
	if !found {
		return nil, true
	}

	if atomic.CompareAndSwapInt64(&h.id, -1, id) {
		return h, false
	}

	return nil, true
}

func (hc *handleServiceImpl) GetHandleWithTuple(nw types.Network, tuple model.Tuple) types.ReteHandle {
	hc.Lock()
	defer hc.Unlock()
	h, found := hc.allHandles[tuple.GetKey().String()]
	if !found {
		h = newReteHandleImpl(nw, tuple, types.ReteHandleStatusCreating, 0)

		hc.allHandles[tuple.GetKey().String()] = h //[tuple.GetKey().String()] = h
	}
	return h
}