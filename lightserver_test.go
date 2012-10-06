package main

import (
	"testing"
	"time"
)

type Time struct {
	Day  int
	Hour int
}

type lsTest struct {
	now          Time
	action       Action
	expectedTime Time
}

var testData = []lsTest{
	{Time{4, 12}, TurnOn, Time{4, 18}},
	{Time{4, 19}, TurnOff, Time{4, 22}},
	{Time{4, 23}, TurnOn, Time{5, 18}},
}

func ToTime(t Time) time.Time {
	return time.Date(2012, 10, t.Day, t.Hour, 0, 0, 0, time.Local)
}

func TestAll(t *testing.T) {
	for _, v := range testData {
		now := ToTime(v.now)
		a, nextTime := nextActionTime(now)
		if a != v.action {
			t.Errorf("Unexpected action for time: %s", now)
		}
		if nextTime.Hour() != v.expectedTime.Hour {
			t.Errorf("Unexpected hour: %d for time: %s", nextTime.Hour(), now)
		}
		if nextTime.Day() != v.expectedTime.Day {
			t.Errorf("Unexpected day: %d for time: %s", nextTime.Day(), now)
		}
	}
}

var testScheduleData = []struct {
	now       Time
	schedule  []ScheduledAction
	expAction Action
	expTime   Time
}{
/*{Time{4, 12}, []ScheduledAction{ ScheduledAction{TurnOn, "1", "18:00"} }, TurnOn, Time{4,18}},*/
}

func TestNextActionAfter(t *testing.T) {
	for _, v := range testScheduleData {
		now := ToTime(v.now)
		a, nextTime := nextActionAfter(now, v.schedule)
		if a != v.expAction {
			t.Errorf("Unexpected action for time: %s", v.now)
		}
		if nextTime.Hour() != v.expTime.Hour {
			t.Errorf("Unexpected hour: %d for time: %s", nextTime.Hour(), now)
		}
		if nextTime.Day() != v.expTime.Day {
			t.Errorf("Unexpected day: %d for time: %s", nextTime.Day(), now)
		}
	}

}

func TestEventsForWeekday(t *testing.T) {
	schedule := ScheduledAction{TurnOn, "1", "18:00"}
}
