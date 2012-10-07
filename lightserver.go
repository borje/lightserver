package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/cpucycle/astrotime"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"sort"
)

const (
	OneDay    = time.Hour * 24
	LATITUDE  = 58.410807
	LONGITUDE = -15.6213727
	/*turnOnBeforeSunset = -time.Hour*/
	turnOnBeforeSunset = 0
	NIGHT_OFF_HOUR     = 22
	NIGHT_OFF_MINUTE   = 36
	/*NIGHT_OFF_HOUR     = 22*/
	/*NIGHT_OFF_MINUTE   = 05*/
)

type Action int

const (
	TurnOn Action = iota
	TurnOff
)

type LightStatus struct {
	Id    int
	Name  string
	State Action
}

func (a Action) String() (s string) {
	if a == TurnOn {
		s = "ON"
	} else {
		s = "OFF"
	}
	return
}

type ScheduleConfigItem struct {
	action   Action
	weekdays string
	time     string
}

type ScheduledEvent struct {
	action Action
	time   time.Time
}

type ScheduledEvents []ScheduledEvent

func (se ScheduledEvents) Len() int { return len(se) }
func (se ScheduledEvents) Swap(i,j int) { se[i], se[j] = se[j], se[i] }
func (se ScheduledEvents) Less(i, j int) bool { return se[i].time.Before(se[j].time)}

func eventsForDay(now time.Time, schedule []ScheduleConfigItem) (events ScheduledEvents) {
	events = make(ScheduledEvents, 0, 7)
	weekDayToSelect := now.Weekday()
	for _, v := range schedule {
		weekdays := strings.Split(v.weekdays, ",")
		for _, wdStr := range weekdays {
			wd, _ := strconv.Atoi(wdStr)
			if weekDayToSelect == time.Weekday(wd) {
				timeStr := strings.Split(v.time, ":")
				hour, _ := strconv.Atoi(timeStr[0])
				minute, _ := strconv.Atoi(timeStr[1])
				newPos := len(events)
				events = events[0 : newPos+1]
				events[newPos] = ScheduledEvent{v.action, time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())}
			}
		}
	}
	sort.Sort(events)
	return
}

func nextActionAfter(now time.Time, schedule []ScheduleConfigItem) (a Action, t time.Time) {
	/*scheduleItem := schedule[0]*/
	/*weekdays := strings.Split(scheduleItem.weekdays,",")*/
	/*var today []ScheduledEvent*/
	/*for i := range weekdays {*/
	/*fmt.Println(weekdays[i])*/
	/*}*/
	return
}

func nextActionTime(now time.Time) (a Action, t time.Time) {
	nightOnTime := astrotime.CalcSunset(now, LATITUDE, LONGITUDE).Add(turnOnBeforeSunset)
	nightOffTime := time.Date(now.Year(), now.Month(), now.Day(), NIGHT_OFF_HOUR, NIGHT_OFF_MINUTE, 0, 0, now.Location())
	if now.Before(nightOnTime) {
		a = TurnOn
		t = nightOnTime
	} else if now.Before(nightOffTime) {
		a = TurnOff
		t = nightOffTime
	} else {
		a = TurnOn
		t = nightOnTime.Add(OneDay)
	}
	return
}

func doTellstickAction(action Action) {
	var tellstickCmd string
	if action == TurnOn {
		tellstickCmd = "--on"
	} else {
		tellstickCmd = "--off"
	}
	cmd := exec.Command("tdtool", tellstickCmd, "2")
	b, err := cmd.CombinedOutput()
	log.Println("Turning device:", action)
	if err != nil {
		log.Fatal("Error executing Tellstick action: ", cmd)
	}
	log.Printf("%s", b)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("tdtool", "--list")
	stdout, _ := cmd.StdoutPipe()
	reader := bufio.NewReader(stdout)
	cmd.Start()
	str, _ := reader.ReadString('\n')
	fields := strings.Fields(str)
	var nrReceivers int
	if len(fields) >= 3 {
		nrReceivers, _ = strconv.Atoi(fields[3])
	} else {
		nrReceivers = 0
	}
	jsonWriter := json.NewEncoder(w)
	for i := 1; i <= nrReceivers; i++ {
		t := &LightStatus{i, "asdf", TurnOff}
		jsonWriter.Encode(t)
	}
}

func schedule(quit chan bool) {
	log.Println("Scheduling started")
	for {
		now := time.Now()
		action, nextTime := nextActionTime(now)
		log.Printf("Next event: %s @ %s", action, nextTime)
		untilNextAction := nextTime.Sub(now)
		timer := time.NewTimer(untilNextAction)
		select {
		case <-quit:
			log.Println("Quit scheduling")
			return
		case <-timer.C:
			log.Println("Executing action: ", action)
			doTellstickAction(action)
		}
	}
}

func main() {
	now := time.Now()
	action, nextTime := nextActionTime(now)
	fmt.Println(action, nextTime)
	if action == TurnOff {
		go doTellstickAction(TurnOn)
	} else {
		go doTellstickAction(TurnOff)
	}
	quit := make(chan bool)
	go schedule(quit)
	time.Sleep(time.Second)
	//quit <- true
	time.Sleep(time.Second)
	http.HandleFunc("/status", statusHandler)
	http.ListenAndServe(":8081", nil)
}
