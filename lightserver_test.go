package main

import (
	"container/heap"
	"fmt"
	"testing"
	"time"
)

func TestParseResponse(t *testing.T) {
	//theDay := time.Now()
	//tm := time.Date(theDay.Year(), theDay.Month(), theDay.Day(), 19, 10, 0, 0, theDay.Location())
	h := &ScheduledEvents{}
	heap.Init(h)
	for i := 0; i < 10; i++ {
		tm := time.Now()
		e := ScheduledEvent{1, TurnOn, tm}
		heap.Push(h, e)
		time.Sleep(time.Second / 5)
	}

	fmt.Println("Len ", h.Len())

	for i, e := range *h {
		fmt.Printf("%d %#v\n", i, e)
		//fmt.Println(e.time, heap.)
	}

	fmt.Println("========")

	for h.Len() > 0 {
		e := heap.Pop(h).(ScheduledEvent)
		fmt.Println(e.time)
	}
}

func TestAsdf(t *testing.T) {

}
