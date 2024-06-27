package main

import (
	"testing"
    "fmt"
)

func TestLimit(t *testing.T){
	l := NewLimit(10_100)
	buyOrder := NewOrder(true, 5)
	buyOrder := NewOrder(true , 8)
	buyOrder := NewOrder(true, 10)
    
	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	fmt.Println(1)
}
func TestOrderbook(t *testing.T) {

}