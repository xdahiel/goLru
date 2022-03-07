package goLru

import (
	"fmt"
	"testing"
	"time"
)

var lifespan = 2 * time.Second

func TestLRUCache(t *testing.T) {
	c := NewLRUCache(3)
	c.Put(1, 10, lifespan)
	c.Put(2, 20, lifespan)
	c.Put(3, 30, lifespan)
	c.Put(4, 40, lifespan)
	fmt.Println(c.Get(3))
	for i := c.items.Front(); i.key != nil; i = i.next {
		fmt.Println(i.key)
	}
	fmt.Println(c.items.Back().key, " ", c.items.Front().key)
	time.Sleep(3 * time.Second)
	fmt.Println(c.Get(2))
	fmt.Println(c.Count())
}
