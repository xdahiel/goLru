package goLru

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	cap  int
	lock sync.Mutex
	list *list.List
	mp   map[interface{}]*list.Element
}

// NewLRUCache construct a new cache
func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		cap:  cap,
		list: list.New(),
		mp:   make(map[interface{}]*list.Element, cap),
	}
}

// Put store a key-value
func (c *LRUCache) Put(k, v interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.list.Len() == c.cap {
		e := c.list.Back()
		c.list.Remove(e)
		delete(c.mp, e.Value)
	}

	fr := c.list.PushFront(v)
	c.mp[k] = fr
}

// Get return the responding value of key
func (c *LRUCache) Get(k interface{}) (interface{}, bool) {
	v, ok := c.mp[k]
	if !ok {
		return nil, false
	} else {
		c.list.MoveToFront(v)
		return v, true
	}
}
