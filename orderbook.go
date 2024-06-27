package main

import (
	"time"
    "fmt"
)

type Order struct {
	Size  float64
	Bid bool
	Limit *Limit
	Timestamp int64
}

func NewOrder(bid bool, size float64)*Order{
	return &Order{
		Size: size,
		Bid : bid,
		Timestamp : time.Now().UnixNano(),
	}
}

func (o *Order) string{
	return fmt.Sprintf(*[size: %.2f])
}

type Limit struct{
 Price  float64
 Orders []*Order 
 TotalVolume float64
}

func NewLimit(price float64) *Limit{
	return &Limit{
		Price: price,
		Orders: []*Order{},
	}
}

func (l *Limit) AddOrder(o *Order){
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}


func(l *Limit) DeleteOrder(o *Order){
	for i:=0; i<len(l.Orders; i++ {
		if l.Orders[i] == o {
			L.Orders[i] = l.Order[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}
	o.Limit = nil
	l.TotalVolume -= o.Size

	/TODO: resort the orders while resting orders
	

}
type Orderbook struct{
	Asks []*Limit 
	Bids []*Limit
}