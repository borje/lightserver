package main

import (
	"lightserver/scheduler"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/unrolled/render"
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
	log.SetFlags(log.Lshortfile)
	c := `[
	{
		"device":2,
		"weekdays":[1,2,3,4,5,6,0],
		"timeFrom":"06:15",
		"timeTo":"21:00"
	},
	{
		"device":2,
		"weekdays":[1,2,3,4,5,6,0],
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

func BenchmarkTesting(b *testing.B) {
	req, err := http.NewRequest("GET", "/schedule?start=2015-09-05&end=2015-10-12&_=1444335823457", nil)
	rend = render.New(render.Options{IndentJSON: true})
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		scheduleHandler(w, req)
		//fmt.Printf("%d - %s", w.Code, w.Body.String())
	}
}
