package main

import (
	/*"bufio"*/
	/*"encoding/json"*/
	/*"github.com/cpucycle/astrotime"*/
	"log"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
	"os/signal"
	"syscall"
)

const (
	OneDay             = time.Hour * 24
	LATITUDE           = 58.410807
	LONGITUDE          = -15.6213727
	LOG_FILE           = "lightserver.log"
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

func (a Action) String() string {
	if a == TurnOn {
		return "ON"
	} else {
		return "OFF"
	}
	return ""
}

type ScheduleConfigItem struct {
	device   int
	weekdays string
	timeFrom string
	timeTo	 string
}

type ScheduledEvent struct {
	device int
	action Action
	time   time.Time
}

type ScheduledEvents []ScheduledEvent

func (se ScheduledEvents) Len() int           { return len(se) }
func (se ScheduledEvents) Swap(i, j int)      { se[i], se[j] = se[j], se[i] }
func (se ScheduledEvents) Less(i, j int) bool { return se[i].time.Before(se[j].time) }

func timeFromString(now time.Time, clock string) (time.Time) {
	timeStr := strings.Split(clock, ":")
	hour, _ := strconv.Atoi(timeStr[0])
	minute, _ := strconv.Atoi(timeStr[1])
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
}

func eventsForDay(now time.Time, schedule []ScheduleConfigItem) (events ScheduledEvents) {
	events = make(ScheduledEvents, 0, 8)
	weekDayToSelect := now.Weekday()
	for _, v := range schedule {
		weekdays := strings.Split(v.weekdays, ",")
		for _, wdStr := range weekdays {
			wd, _ := strconv.Atoi(wdStr)
			if weekDayToSelect == time.Weekday(wd) {
				var hour int
				var minute int
				/*if v.timeFrom == "SUNSET" {*/
					/*sunset := astrotime.CalcSunset(now, LATITUDE, LONGITUDE)*/
					/*hour = sunset.Hour()*/
					/*minute = sunset.Minute()*/
				/*} else if v.time == "SUNRISE" {*/
					/*sunset := astrotime.CalcSunrise(now, LATITUDE, LONGITUDE)*/
					/*hour = sunset.Hour()*/
					/*minute = sunset.Minute()*/
				/*} else if v.time[2] == ':' {*/

				// TURN ON
				if v.timeFrom[2] == ':'{
					timeStr := strings.Split(v.timeFrom, ":")
					hour, _ = strconv.Atoi(timeStr[0])
					minute, _ = strconv.Atoi(timeStr[1])
				}
				newPos := len(events)
				events = events[0 : newPos+1]
				events[newPos] = ScheduledEvent{v.device, TurnOn, time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())}

				// TURN OFF
				if v.timeTo[2] == ':'{
					timeStr := strings.Split(v.timeTo, ":")
					hour, _ = strconv.Atoi(timeStr[0])
					minute, _ = strconv.Atoi(timeStr[1])
				}
				newPos = len(events)
				events = events[0 : newPos+1]
				events[newPos] = ScheduledEvent{v.device, TurnOff, time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())}
			}
		}
	}
	sort.Sort(events)
	return
}

func nextActionAfter(now time.Time, schedule []ScheduleConfigItem) (int, Action, time.Time) {
	for {
		for _, event := range eventsForDay(now, schedule) {
			if event.time.After(now) {
				return event.device, event.action, event.time
			}
		}
		nextDay := now.Add(OneDay)
		now = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())
	}
	log.Fatal("Should not return here")
	return 0, TurnOn, now
}

func doTellstickAction(action Action) {
	var tellstickCmd string
	if action == TurnOn {
		tellstickCmd = "--on"
	} else {
		tellstickCmd = "--off"
	}
	/*cmd := exec.Command("tdtool", tellstickCmd, "2")*/
	cmd := exec.Command("echo", tellstickCmd, "2")
	b, err := cmd.CombinedOutput()
	log.Println("Turning device:", action)
	if err != nil {
		log.Println("Error executing Tellstick action: ", cmd)
		// Some kind error handling
	}
	log.Printf("%s", b)
}

/*
  Web interface
*/

func statusHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	for i := 0; i < 4 * 7; i++ {
		device, action, next := nextActionAfter(now, GetConfiguration())
		fmt.Println(device, next, action)
		now = next
	}
}

func schedule(configuration []ScheduleConfigItem, quit chan bool) {
	for {
		now := time.Now()
		_, action, nextTime := nextActionAfter(now, configuration)
		log.Printf("Next event: %s @ %s (device %d)", action, nextTime)
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

func signalHandler(quit chan bool) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChannel
	quit <- true
	log.Println("Received signal: ", sig)
}

func GetConfiguration() []ScheduleConfigItem {
	return []ScheduleConfigItem{
		{2, "1,2,3,4,5,6,0", "15:00", "22:15"},
		{2, "1,2,3,4,5,6,0", "07:15", "09:30"},
		{1, "1,2,3,4,5,6,0", "05:35", "11:00"},
		{1, "1,2,3,4,5,6,0", "13:00", "22:15"},
	}
}

func main() {
	configuration := GetConfiguration()
	logfile, err := os.OpenFile(LOG_FILE, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logfile)
		defer func () {
			log.Println("Exiting")
			logfile.Close()
		} ()
	}
	log.Println("Starting")
	now := time.Now()
	for i := 0; i < 8; i = i + 1 {
		device, action, next := nextActionAfter(now, configuration)
		fmt.Println(device, next, action)
		now = next
	}
	/// Set correct light status at startup
	_, action, _ := nextActionAfter(now, configuration)
	if action == TurnOff {
		go doTellstickAction(TurnOn)
	} else {
		go doTellstickAction(TurnOff)
	}
	quit := make(chan bool)
	go schedule(configuration, quit)
	http.HandleFunc("/status", statusHandler)
	go http.ListenAndServe(":8081", nil)
	signalHandler(quit)
}
