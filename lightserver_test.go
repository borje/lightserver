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
	schedule  []ScheduleConfigItem
	expAction Action
	expTime   Time
}{
	{Time{1, 12}, []ScheduleConfigItem{ScheduleConfigItem{TurnOn, "1", "18:00"}}, TurnOn, Time{1, 18}},
	{Time{9, 18}, []ScheduleConfigItem{ScheduleConfigItem{TurnOff, "2", "21:00"}}, TurnOff, Time{9, 21}},
}

func aTestNextActionAfter(t *testing.T) {
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
			t.Errorf("Unexpected day for time: %s", nextTime.Day(), now)
		}
	}

}

func TestEventsForWeekday(t *testing.T) {
	for _, testData := range testScheduleData {
		schedule := testData.schedule
		now := ToTime(testData.now)
		events := eventsForDay(now, schedule)
		if len(events) != 1 {
			t.Errorf("Exptected length: %d, received: %d", 1, len(events))
		}
		for _, e := range events {
			if e.action != testData.expAction {
				t.Errorf("Unexpected action for time: %s. Got %s, expected %s", now, e.action, TurnOn)
			}
			if e.time.Hour() != testData.expTime.Hour {
				t.Errorf("Unexpected hour: %2 for time: %s", e.time.Hour(), now)
			}
			if e.time.Day() != testData.expTime.Day {
				t.Errorf("Unexpected day: %d for time: %s", e.time.Day(), now)
			}
		}
	}
}
