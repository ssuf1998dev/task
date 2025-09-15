package ast

import (
	"iter"
	"maps"
	"sync"

	"github.com/elliotchance/orderedmap/v3"

	"github.com/go-task/task/v3/internal/deepcopy"
)

type (
	Store struct {
		om    *orderedmap.OrderedMap[string, any]
		mutex sync.RWMutex
	}
	StoreElement orderedmap.Element[string, any]
)

func NewStore(els ...*StoreElement) *Store {
	store := &Store{
		om: orderedmap.NewOrderedMap[string, any](),
	}
	for _, el := range els {
		store.Set(el.Key, el.Value)
	}
	return store
}

func (s *Store) Len() int {
	if s == nil || s.om == nil {
		return 0
	}
	defer s.mutex.RUnlock()
	s.mutex.RLock()
	return s.om.Len()
}

func (s *Store) Get(key string) (any, bool) {
	if s == nil || s.om == nil {
		return nil, false
	}
	defer s.mutex.RUnlock()
	s.mutex.RLock()
	return s.om.Get(key)
}

func (s *Store) Set(key string, value any) bool {
	if s == nil {
		s = NewStore()
	}
	if s.om == nil {
		s.om = orderedmap.NewOrderedMap[string, any]()
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	return s.om.Set(key, value)
}

func (s *Store) All() iter.Seq2[string, any] {
	if s == nil || s.om == nil {
		return func(yield func(string, any) bool) {}
	}
	return s.om.AllFromFront()
}

func (s *Store) Keys() iter.Seq[string] {
	if s == nil || s.om == nil {
		return func(yield func(string) bool) {}
	}
	return s.om.Keys()
}

func (s *Store) Values() iter.Seq[any] {
	if s == nil || s.om == nil {
		return func(yield func(any) bool) {}
	}
	return s.om.Values()
}

func (s *Store) ToMap() (m map[string]any) {
	defer s.mutex.RUnlock()
	s.mutex.RLock()
	m = maps.Collect(s.All())
	return m
}

func (s *Store) DeepCopy() *Store {
	if s == nil {
		return nil
	}
	defer s.mutex.RUnlock()
	s.mutex.RLock()
	return &Store{
		om: deepcopy.OrderedMap(s.om),
	}
}
