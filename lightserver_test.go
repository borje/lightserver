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

type AT struct {
	action Action
	time   Time
}

var testScheduleData = []struct {
	now      Time
	schedule []ScheduleConfigItem
	expected []AT
}{
	{Time{1, 12},
		[]ScheduleConfigItem{{TurnOn, "1", "18:00"}},
		[]AT{{TurnOn, Time{1, 18}}}},
	{Time{9, 18},
		[]ScheduleConfigItem{
			{TurnOff, "2", "21:00"},
			{TurnOn,  "2", "22:00"},
		},
		[]AT{
			{TurnOff, Time{9, 21}},
			{TurnOn, Time{9, 22}},
		}},
	{Time{9, 18},
		[]ScheduleConfigItem{
			{TurnOn,  "2", "22:00"},
			{TurnOff, "2", "21:00"},
		},
		[]AT{
			{TurnOff, Time{9, 21}},
			{TurnOn, Time{9, 22}},
		}},
}

func aTestNextActionAfter(t *testing.T) {
	for _, v := range testScheduleData {
		now := ToTime(v.now)
		a, nextTime := nextActionAfter(now, v.schedule)
		if a != v.expected[0].action {
			t.Errorf("Unexpected action for time: %s", v.now)
		}
		if nextTime.Hour() != v.expected[0].time.Hour {
			t.Errorf("Unexpected hour: %d for time: %s", nextTime.Hour(), now)
		}
		if nextTime.Day() != v.expected[0].time.Day {
			t.Errorf("Unexpected day for time: %s", nextTime.Day(), now)
		}
	}
}

func TestEventsForWeekday(t *testing.T) {
	for _, testData := range testScheduleData {
		schedule := testData.schedule
		now := ToTime(testData.now)
		events := eventsForDay(now, schedule)
		if len(events) != len(testData.expected) {
			t.Errorf("Exptected length: %d, received: %d", 1, len(events))
		}
		for i, e := range events {
			if e.action != testData.expected[i].action {
				t.Errorf("Unexpected action for time: %s. Got %s, expected %s", now, e.action, TurnOn)
			}
			if e.time.Hour() != testData.expected[i].time.Hour {
				t.Errorf("Unexpected hour: %d for time: %s", e.time.Hour(), now)
			}
			if e.time.Day() != testData.expected[i].time.Day {
				t.Errorf("Unexpected day: %d for time: %s", e.time.Day(), now)
			}
		}
	}
}
