package goLru

import (
	"sync"
	"time"
)

// CacheItem the element of a unidirectional linked list
type CacheItem struct {
	sync.RWMutex

	// item's key
	key interface{}
	// item's value
	value interface{}

	// item's lifespan
	lifespan time.Duration
	// when the item is created
	createdAt time.Time
	// how much time the item is accessed
	accessCount int64
	// the last time of access
	accessAt time.Time

	// the items are stored in a unidirectional linked list
	next, prev *CacheItem
	// the list to which this element belongs
	list *linkedList
}

// NewCacheItem create a new CacheItem
func NewCacheItem(key, value interface{}, lifespan time.Duration) *CacheItem {
	now := time.Now()
	return &CacheItem{
		key:         key,
		value:       value,
		lifespan:    lifespan,
		createdAt:   now,
		accessAt:    now,
		accessCount: 0,
	}
}

// activate the item
func (item *CacheItem) keepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessAt = time.Now()
	item.accessCount++
}

func (item *CacheItem) Key() interface{} {
	return item.key
}

func (item *CacheItem) SetValue(v interface{}) {
	item.value = v
}

func (item *CacheItem) Value() interface{} {
	return item.value
}

func (item *CacheItem) Lifespan() time.Duration {
	return item.lifespan
}

func (item *CacheItem) CreatedAt() time.Time {
	return item.createdAt
}

func (item *CacheItem) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

func (item *CacheItem) AccessTime() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessAt
}

func (item *CacheItem) Next() *CacheItem {
	if p := item.next; item.list != nil && p != &item.list.root {
		return p
	}
	return nil
}

func (item *CacheItem) Prev() *CacheItem {
	if p := item.prev; item.list != nil && p != &item.list.root {
		return p
	}
	return nil
}

type linkedList struct {
	root CacheItem
	len  int
}

func (l *linkedList) Init() *linkedList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

func NewLinkedList() *linkedList {
	return new(linkedList).Init()
}

func (l *linkedList) Len() int {
	return l.len
}

func (l *linkedList) Front() *CacheItem {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *linkedList) Back() *CacheItem {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *linkedList) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

func (l *linkedList) insert(e, at *CacheItem) *CacheItem {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

func (l *linkedList) remove(e *CacheItem) *CacheItem {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
	return e
}

func (l *linkedList) move(e, at *CacheItem) *CacheItem {
	if e == at {
		return e
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e

	return e
}

func (l *linkedList) Remove(e *CacheItem) interface{} {
	if e.list == l {
		l.remove(e)
	}
	return e.value
}

func (l *linkedList) PushFront(v *CacheItem) *CacheItem {
	l.lazyInit()
	return l.insert(v, &l.root)
}

func (l *linkedList) MoveToFront(e *CacheItem) {
	if e.list != l || l.root.next == e {
		return
	}
	l.move(e, &l.root)
}
