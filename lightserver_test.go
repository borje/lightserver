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

func ToTime(t Time) time.Time {
	return time.Date(2012, 10, t.Day, t.Hour, 0, 0, 0, time.Local)
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
			{TurnOn, "2", "22:00"},
		},
		[]AT{
			{TurnOff, Time{9, 21}},
			{TurnOn, Time{9, 22}},
		}},
	{Time{9, 18},
		[]ScheduleConfigItem{
			{TurnOn, "2", "22:00"},
			{TurnOff, "2", "21:00"},
		},
		[]AT{
			{TurnOff, Time{9, 21}},
			{TurnOn, Time{9, 22}},
		}},
}

var testData2 = []struct {
	now            Time
	config         []ScheduleConfigItem
	expectedAction Action
	expectedTime   Time
}{
	{Time{8, 15},
		[]ScheduleConfigItem{
			{TurnOn, "1", "18:00"}},
		TurnOn,
		Time{8, 18}},
	{Time{8, 19},
		[]ScheduleConfigItem{
			{TurnOn, "1", "18:00"},
			{TurnOn, "1", "20:00"}},
		TurnOn,
		Time{8, 20}},
	{Time{8, 21},
		[]ScheduleConfigItem{
			{TurnOn, "1", "20:00"},
			{TurnOn, "1,2", "17:00"}},
		TurnOn,
		Time{9, 17}},
	{Time{9, 3},
		[]ScheduleConfigItem{
			{TurnOn, "2", "SUNSET"},
			{TurnOn, "2", "SUNRISE"}},
		TurnOn,
		Time{9, 7}},
	{Time{9, 3},
		[]ScheduleConfigItem{
			{TurnOn, "2", "07:30"},
			{TurnOff, "2", "SUNRISE"},
			{TurnOn, "2", "17:00"}},
		TurnOn,
		Time{9, 17}},
}

func TestNextActionAfter(t *testing.T) {
	for _, v := range testData2 {
		now := ToTime(v.now)
		action, nextTime := nextActionAfter(now, v.config)
		if action != v.expectedAction {
			t.Errorf("Unexpected action for time: %s", v.now)
		}
		if nextTime.Hour() != v.expectedTime.Hour {
			t.Errorf("Unexpected hour: %d for time: %s", nextTime.Hour(), now)
		}
		if nextTime.Day() != v.expectedTime.Day {
			t.Errorf("Unexpected day: %d for time: %s", nextTime.Day(), now)
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
