package goLru

import (
	"sync"
	"time"
)

const DefaultLifeSpan = 180 * 24 * time.Hour

type LRUCache struct {
	sync.RWMutex

	// the capacity of cache
	cap int
	// unidirectional list store the items
	items *linkedList
	// hash map store the key/value pair
	mp map[interface{}]*CacheItem

	// timer responsible for triggering cleanup
	cleanupTimer *time.Timer
	// current timer duration
	cleanupInterval time.Duration
}

// NewLRUCache return a new LRUCache
func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		cap:   cap,
		items: NewLinkedList(),
		mp:    make(map[interface{}]*CacheItem, cap),
	}
}

// Tail return the tail of linked list
func (c *LRUCache) Tail() *CacheItem {
	return c.items.Back()
}

// Count return the count of items in the cache
func (c *LRUCache) Count() int {
	c.RLock()
	defer c.RUnlock()

	return c.items.Len()
}

// OperateForAll do something for all items
func (c *LRUCache) OperateForAll(op func(key interface{}, item *CacheItem)) {
	c.RLock()
	defer c.RUnlock()

	for k, v := range c.mp {
		op(k, v)
	}
}

// Add a key/value pair
func (c *LRUCache) add(k, v interface{}, lifespan time.Duration) *CacheItem {
	item := NewCacheItem(k, v, lifespan)

	c.Lock()
	c.addInternal(item)
	return item
}

func (c *LRUCache) addInternal(item *CacheItem) {
	c.mp[item.key] = item
	c.items.PushFront(item)

	expDur := c.cleanupInterval
	c.Unlock()

	//fmt.Println(item.key, expDur)

	if item.lifespan > 0 && (expDur == 0 || item.lifespan < expDur) {
		c.expirationCheck()
	}
	//go c.expirationCheck()
}

func (c *LRUCache) deleteInternal(key interface{}) (*CacheItem, error) {
	r, ok := c.mp[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	c.Unlock()

	c.Lock()
	c.items.Remove(r)
	delete(c.mp, key)

	return r, nil
}

// Delete remove a key/value pair according to given key, if success, return removed item
func (c *LRUCache) Delete(key interface{}) (*CacheItem, error) {
	c.Lock()
	defer c.Unlock()

	return c.deleteInternal(key)
}

// expiration check, triggered by a self-adjusting timer.
func (c *LRUCache) expirationCheck() {
	c.Lock()
	if c.cleanupTimer != nil {
		c.cleanupTimer.Stop()
	}

	now := time.Now()
	smallestDuration := 0 * time.Second

	//fmt.Println(gap, item.lifespan, item.key, "-------")

	// find a minimal duration to do a check again
	for key, item := range c.mp {
		item.RLock()
		lifespan := item.lifespan
		accessedAt := item.accessAt
		item.RUnlock()

		if lifespan == 0 {
			continue
		}

		if now.Sub(accessedAt) >= lifespan {
			c.deleteInternal(key)
		} else {
			if smallestDuration == 0 || lifespan-now.Sub(accessedAt) < smallestDuration {
				smallestDuration = lifespan - now.Sub(accessedAt)
			}
			break
		}
	}

	c.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		c.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go c.expirationCheck()
		})
	}
	c.Unlock()
}

// Exists judge if given key is exists
func (c *LRUCache) Exists(key interface{}) bool {
	c.RLock()
	defer c.RUnlock()

	_, ok := c.mp[key]
	return ok
}

// Put Update the corresponding item's value or add a key/value pair
func (c *LRUCache) Put(key, value interface{}, lifespan time.Duration) *CacheItem {
	if !c.Exists(key) {
		if c.Count() == c.cap {
			c.Delete(c.Tail().key)
		}
		return c.add(key, value, lifespan)
	}
	c.Lock()
	defer c.Unlock()

	item := c.mp[key]
	item.SetValue(value)
	item.keepAlive()
	c.items.MoveToFront(item)

	return item
}

// Get value according to key
func (c *LRUCache) Get(key interface{}, args ...interface{}) (*CacheItem, error) {
	c.RLock()
	r, ok := c.mp[key]
	c.RUnlock()

	if ok {
		r.keepAlive()
		c.Lock()
		c.items.MoveToFront(r)
		c.Unlock()
		return r, nil
	}

	return nil, ErrKeyNotFound
}

// Flush remove all data
func (c *LRUCache) Flush() {
	c.Lock()
	defer c.Unlock()

	c.mp = make(map[interface{}]*CacheItem)
	c.items = c.items.Init()
	c.cleanupInterval = 0
	if c.cleanupTimer != nil {
		c.cleanupTimer.Stop()
	}
}
