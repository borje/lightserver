package main

import (
	"lightserver/scheduler"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestShouldBeSortedWhenSameObjects(t *testing.T) {
	events := scheduler.ScheduledEvents{}
	events = append(events, scheduler.ScheduledEvent{2, scheduler.TurnOn, time.Date(2015, 9, 12, 6, 15, 0, 0, time.Local)})
	if !sort.IsSorted(events) {
		t.Errorf("Not sorted")
	}
	events = append(events, scheduler.ScheduledEvent{2, scheduler.TurnOn, time.Date(2015, 9, 12, 6, 15, 0, 0, time.Local)})
	sort.Sort(events)
	if !sort.IsSorted(events) {
		t.Errorf("Not sorted")
	}
}

func TestReturnCopyOfEventQueue(t *testing.T) {
	c := `[
	{
		"device":2,
		"weekdays":"1,2,3,4,5,6,0",
		"timeFrom":"06:15",
		"timeTo":"21:00"
	},
	{
		"device":2,
		"weekdays":"1,2,3,4,5,6,0",
		"timeFrom":"06:15",
		"timeTo":"20:00"
	}
	]
	`

	s := scheduler.NewSchedulerFromReader(strings.NewReader(c))
	day, _ := time.Parse("2006-01-02", "2015-09-12")
	s.AddEventsForDay(day)
	queue1 := s.EventQueue()
	queue2 := s.EventQueue()
	queue1[0] = scheduler.ScheduledEvent{}

	if queue2[0] == queue1[0] {
		t.Errorf("Should return copy of the event queue")
	}
	if !sort.IsSorted(queue2) {
		t.Errorf("Should be sorted")
	}
}
