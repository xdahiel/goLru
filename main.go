package goLru

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	cap    int
	expire int
	lock   sync.Mutex
	list   *list.List
	mp     map[interface{}]*list.Element
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		cap:  cap,
		list: list.New(),
		mp:   make(map[interface{}]*list.Element, cap),
	}
}

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

func (c *LRUCache) Get(k interface{}) (interface{}, bool) {
	v, ok := c.mp[k]
	if !ok {
		return nil, false
	} else {
		c.list.MoveToFront(v)
		return v, true
	}
}
