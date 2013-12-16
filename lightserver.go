package main

import (
	/*"bufio"*/
	/*"encoding/json"*/
	/*"github.com/cpucycle/astrotime"*/
	"container/heap"
	"log"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"os/signal"
	"syscall"
)

const (
	LATITUDE           = 58.410807
	LONGITUDE          = -15.6213727
	LOG_FILE           = "lightserver.log"
)

type Action int

const (
	TurnOn Action = iota
	TurnOff
)

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

var eventQueue *ScheduledEvents

/* Functions for sorting ScheduledEvents */
func (se ScheduledEvents) Len() int           { return len(se) }
func (se ScheduledEvents) Swap(i, j int)      { se[i], se[j] = se[j], se[i] }
func (se ScheduledEvents) Less(i, j int) bool { return !se[i].time.After(se[j].time) }

/* heap functions */
func (se *ScheduledEvents) Pop() interface{} {
	old := *se
	n := len(old)
	x := old[n-1]
	*se = old[0 : n-1]
	return x
}

func (se *ScheduledEvents) Push(x interface{}) {
	*se = append(*se, x.(ScheduledEvent))
}

func timeFromString(theDay time.Time, clock string) (time.Time) {
	timeStr := strings.Split(clock, ":")
	hour, _ := strconv.Atoi(timeStr[0])
	minute, _ := strconv.Atoi(timeStr[1])
	return time.Date(theDay.Year(), theDay.Month(), theDay.Day(), hour, minute, 0, 0, theDay.Location())
}

func eventsForDay(now time.Time, schedule []ScheduleConfigItem) (events ScheduledEvents) {
	const MAX_EVENTS_PER_DAY_PER_DEVICE = 8
	events = ScheduledEvents{}
	currentWeekDay := now.Weekday()
	for _, v := range schedule { // unused return value ?
		device := v.device
		weekdays := strings.Split(v.weekdays, ",")
		for _, dayInWeekString := range weekdays {
			dayInWeek, _ := strconv.Atoi(dayInWeekString)
			if currentWeekDay == time.Weekday(dayInWeek) {
				/*var hour int*/
				/*var minute int*/
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
					newEvent := ScheduledEvent{device, TurnOn, timeFromString(now, v.timeFrom)}
					events = append(events, newEvent)
				}

				// TURN OFF
				if v.timeTo[2] == ':'{
					newEvent := ScheduledEvent{device, TurnOff, timeFromString(now, v.timeTo)}
					events = append(events, newEvent)
				}
			}
		}
	}
	return
}

func doTellstickAction(device int, action Action) {
	var tellstickCmd string
	if action == TurnOn {
		tellstickCmd = "--on"
	} else {
		tellstickCmd = "--off"
	}
	cmd := exec.Command("tdtool", tellstickCmd, strconv.Itoa(device))
	b, err := cmd.CombinedOutput()
	log.Println("Turning device", device, ":", action)
	if err != nil {
		log.Println("Error executing Tellstick action: ", cmd)
		// Some kind error handling
	}
	log.Printf("CombinedOutput: %s", b)
}

/*
  Web interface
*/

func statusHandler(w http.ResponseWriter, r *http.Request) {
	for i := eventQueue.Len(); i > 0; i-- {
		e := (*eventQueue)[i-1]
		fmt.Fprintf(w, "%3s %d @ %s\n", e.action, e.device, e.time)
	}
}

func schedule(events *ScheduledEvents, quit chan bool) {
	currentDay := time.Now()
	for {
		for events.Len() > 0 {
			event := heap.Pop(events).(ScheduledEvent)
			log.Printf("Next event: %s @ %s (device %d)", event.action, event.time, event.device)
			if now := time.Now(); now.Before(event.time) {
				log.Printf("Sleeping for %s", event.time.Sub(now))
				timer := time.NewTimer(event.time.Sub(now))
				select {
				case <-quit:
					log.Println("Quit scheduling")
					return
				case <-timer.C:
					doTellstickAction(event.device, event.action)
				}
			} else {
				doTellstickAction(event.device, event.action)
			}
		}
		currentDay = currentDay.AddDate(0, 0, 1)
		addEventForDay(currentDay)
	}
}

func signalHandler(quit chan bool) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChannel
	quit <- true
	log.Println("Received signal: ", sig)
}

func getConfiguration() []ScheduleConfigItem {
	return []ScheduleConfigItem{
		{2, "1,2,3,4,5,0", "07:15", "09:15"},
		{2, "1,2,3,4,5,0", "14:30", "22:15"},
		{1, "1,2,3,4,5,0", "05:30", "22:15"},
	}
}

func configuredDevices(configuration []ScheduleConfigItem) (devices []int) {
	deviceMap := map[int]bool{}
	for _, i := range configuration {
		deviceMap[i.device] = true
	}
	for k := range deviceMap {
		devices = append(devices, k)
	}
	return devices
}

func addEventForDay(day time.Time) {
	for _, event := range eventsForDay(day, getConfiguration()) {
		log.Println("Adding event @", event.time)
		heap.Push(eventQueue, event)
	}
}

func initialState() time.Time {
	log.Println("Initial states")
	now := time.Now()
	currentDay := now
	for _, device := range configuredDevices(getConfiguration()) {
		actionFound := false
		for actionFound == false {
			for _, event := range *eventQueue { // Iterate the help backwards
				if event.device == device && event.time.Before(now) {
					doTellstickAction(device, event.action)
					actionFound = true
					break
				}
			}
			if actionFound == false {
				addEventForDay(currentDay)
				currentDay = currentDay.AddDate(0, 0, -1)
			}
		}
	}
	// remove old events
	for eventQueue.Len() > 0 {
		next := heap.Pop(eventQueue).(ScheduledEvent)
		log.Printf("Removing %d @ %s", next.device, next.time)
		if next.time.After(now) {
			heap.Push(eventQueue, next)
			break
		}
	}
	return now
}

func main() {
	logfile, err := os.OpenFile(LOG_FILE, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logfile)
		defer func () {
			log.Println("Exiting")
			logfile.Close()
		} ()
	}
	log.Println("Starting")
	eventQueue = &ScheduledEvents{}
	heap.Init(eventQueue)
	initialState()

	// DEBUG
	for i := eventQueue.Len(); i > 0; i-- {
		e := (*eventQueue)[i-1]
		fmt.Println(e.time, e.device, e.action)
	}

	quit := make(chan bool)
	go schedule(eventQueue, quit)
	http.HandleFunc("/status", statusHandler)
	go http.ListenAndServe(":8081", nil)
	signalHandler(quit)
}
