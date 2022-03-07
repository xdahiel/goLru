package goLru

import (
	"sync"
	"time"
)

// CacheItem the CacheItem of a unidirectional linked linkedList
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

	// the items are stored in a unidirectional linked linkedList
	next, prev *CacheItem
	// the linkedList to which this CacheItem belongs
	linkedList *linkedList
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
	if p := item.next; item.linkedList != nil && p != &item.linkedList.root {
		return p
	}
	return nil
}

func (item *CacheItem) Prev() *CacheItem {
	if p := item.prev; item.linkedList != nil && p != &item.linkedList.root {
		return p
	}
	return nil
}

type linkedList struct {
	root CacheItem
	len  int
}

func NewLinkedList() *linkedList {
	return new(linkedList).Init()
}

func (l *linkedList) Len() int {
	return l.len
}

func (l *linkedList) Init() *linkedList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// lazyInit lazily initializes a zero linkedList value.
func (l *linkedList) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// Front returns the first element of list l or nil if the list is empty.
func (l *linkedList) Front() *CacheItem {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *linkedList) Back() *CacheItem {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// insert inserts e after at, increments l.len, and returns e.
func (l *linkedList) insert(e, at *CacheItem) *CacheItem {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.linkedList = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&CacheItem{Value: v}, at).
func (l *linkedList) insertValue(v, at *CacheItem) *CacheItem {
	return l.insert(v, at)
}

// remove removes e from its linkedList, decrements l.len, and returns e.
func (l *linkedList) remove(e *CacheItem) *CacheItem {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.linkedList = nil
	l.len--
	return e
}

// move moves e to next to at and returns e.
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

// Remove removes e from l if e is an CacheItem of linkedList l.
// It returns the CacheItem value e.Value.
// The CacheItem must not be nil.
func (l *linkedList) Remove(e *CacheItem) interface{} {
	if e.linkedList == l {
		// if e.linkedList == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero CacheItem) and l.remove will crash
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts a new CacheItem e with value v at the front of linkedList l and returns e.
func (l *linkedList) PushFront(v *CacheItem) {
	l.lazyInit()
	l.insertValue(v, &l.root)
}

// PushBack inserts a new CacheItem e with value v at the back of linkedList l and returns e.
func (l *linkedList) PushBack(v *CacheItem) {
	l.lazyInit()
	l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new CacheItem e with value v immediately before mark and returns e.
// If mark is not an CacheItem of l, the linkedList is not modified.
// The mark must not be nil.
func (l *linkedList) InsertBefore(v, mark *CacheItem) {
	if mark.linkedList != l {
		return
	}
	// see comment in linkedList.Remove about initialization of l
	l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new CacheItem e with value v immediately after mark and returns e.
// If mark is not an CacheItem of l, the linkedList is not modified.
// The mark must not be nil.
func (l *linkedList) InsertAfter(v, mark *CacheItem) {
	if mark.linkedList != l {
		return
	}
	// see comment in linkedList.Remove about initialization of l
	l.insertValue(v, mark)
}

// MoveToFront moves CacheItem e to the front of linkedList l.
// If e is not an CacheItem of l, the linkedList is not modified.
// The CacheItem must not be nil.
func (l *linkedList) MoveToFront(e *CacheItem) {
	if e.linkedList != l || l.root.next == e {
		return
	}
	// see comment in linkedList.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves CacheItem e to the back of linkedList l.
// If e is not an CacheItem of l, the linkedList is not modified.
// The CacheItem must not be nil.
func (l *linkedList) MoveToBack(e *CacheItem) {
	if e.linkedList != l || l.root.prev == e {
		return
	}
	// see comment in linkedList.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves CacheItem e to its new position before mark.
// If e or mark is not an CacheItem of l, or e == mark, the linkedList is not modified.
// The CacheItem and mark must not be nil.
func (l *linkedList) MoveBefore(e, mark *CacheItem) {
	if e.linkedList != l || e == mark || mark.linkedList != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves CacheItem e to its new position after mark.
// If e or mark is not an CacheItem of l, or e == mark, the linkedList is not modified.
// The CacheItem and mark must not be nil.
func (l *linkedList) MoveAfter(e, mark *CacheItem) {
	if e.linkedList != l || e == mark || mark.linkedList != l {
		return
	}
	l.move(e, mark)
}
