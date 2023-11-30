package main

import (
	"sync"
)

type Cache struct {
	mu     sync.Mutex
	Orders []Order
}

// NewCache создает и возвращает новый экземпляр кэша
func NewCache() *Cache {
	return &Cache{
		mu:     sync.Mutex{},
		Orders: make([]Order, 0),
	}
}

func (c *Cache) AddOrders(orders []Order, err2 error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Orders = append(c.Orders, orders...)
}

func (c *Cache) AddOrder(order Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Orders = append(c.Orders, order)
}
