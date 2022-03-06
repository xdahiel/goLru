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
	c.Get(1)
	time.Sleep(3 * time.Second)
	fmt.Println(c.Get(2))
	fmt.Println(c.Count())
	fmt.Println(c.items.Back().key)
}
