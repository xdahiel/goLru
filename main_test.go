package goLru

import (
	"fmt"
	"testing"
)

func TestLRUCache(t *testing.T) {
	c := NewLRUCache(3)
	c.Put(1, 2)
	c.Put(2, 3)
	fmt.Println(c.Get(1))
	fmt.Println(c.Get(3))
	c.Put(3, 4)
	c.Put(4, 2)
	c.Put(5, 100)
	fmt.Println(c.Get(6))
	c.Put(6, 2)
	fmt.Println(c.Get(2))
	fmt.Println(c.Get(6))
}
