package main

import (
	/*"bufio"*/
	"container/heap"
	"encoding/json"
	"fmt"
	"github.com/cpucycle/astrotime"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	LATITUDE  = 58.410807
	LONGITUDE = -15.6213727
	LOG_FILE  = "lightserver.log"
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
	Device   int
	Weekdays string
	TimeFrom string
	TimeTo   string
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

func timeFromString(theDay time.Time, clock string) time.Time {
	timeStr := strings.Split(clock, ":")
	hour, _ := strconv.Atoi(timeStr[0])
	minute, _ := strconv.Atoi(timeStr[1])
	return time.Date(theDay.Year(), theDay.Month(), theDay.Day(), hour, minute, 0, 0, theDay.Location())
}

func eventsForDay(now time.Time, schedule []ScheduleConfigItem) (events ScheduledEvents) {
	events = ScheduledEvents{}
	currentWeekDay := now.Weekday()
	for _, v := range schedule { // unused return value ?
		device := v.Device
		weekdays := strings.Split(v.Weekdays, ",")
		for _, dayInWeekString := range weekdays {
			dayInWeek, _ := strconv.Atoi(dayInWeekString)
			if currentWeekDay == time.Weekday(dayInWeek) {
				// TURN ON
				var onEvent ScheduledEvent
				var offEvent ScheduledEvent
				if v.TimeFrom == "SUNSET" {
					onEvent = ScheduledEvent{device, TurnOn, astrotime.CalcSunset(now, LATITUDE, LONGITUDE)}
				} else if v.TimeFrom[2] == ':' {
					onEvent = ScheduledEvent{device, TurnOn, timeFromString(now, v.TimeFrom)}
				}
				// TURN OFF
				if v.TimeTo == "SUNRISE" {
					offEvent = ScheduledEvent{device, TurnOff, astrotime.CalcSunrise(now, LATITUDE, LONGITUDE)}
				} else if v.TimeTo[2] == ':' {
					offEvent = ScheduledEvent{device, TurnOff, timeFromString(now, v.TimeTo)}
				}
				if onEvent.time.Before(offEvent.time) {
					events = append(events, onEvent)
					events = append(events, offEvent)
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

func logHandler(w http.ResponseWriter, r *http.Request) {
	logfile, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logfile)
		defer func() {
			log.Println("Exiting")
			logfile.Close()
		}()
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
	jsonStream := `
	[
	{"device": 2, "weekdays": "1,2,3,4,5", "timeFrom": "06:15", "timeTo": "SUNRISE"},
	{"device": 2, "weekdays": "6,0", "timeFrom": "07:15", "timeTo": "SUNRISE"},
	{"device": 2, "weekdays": "1,2,3,4,5,6,0", "timeFrom": "SUNSET", "timeTo": "22:15"},
	{"device": 1, "weekdays": "1,2,3,4,5,6,0", "timeFrom": "07:00", "timeTo": "22:15"}
	]`
	jsonDecoder := json.NewDecoder(strings.NewReader(jsonStream))
	var config []ScheduleConfigItem
	err := jsonDecoder.Decode(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	return config
}

func configuredDevices(configuration []ScheduleConfigItem) (devices []int) {
	deviceMap := map[int]bool{}
	for _, i := range configuration {
		deviceMap[i.Device] = true
	}
	for k := range deviceMap {
		devices = append(devices, k)
	}
	return devices
}

func addEventForDay(day time.Time) {
	for _, event := range eventsForDay(day, getConfiguration()) {
		log.Println("Adding ", event.action, " event @", event.time)
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
			for i := range *eventQueue {
				event := (*eventQueue)[eventQueue.Len()-1-i] // Iterate the heap backwards
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
	logfile, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logfile)
		defer func() {
			log.Println("Exiting")
			logfile.Close()
		}()
	}
	log.Println("Starting")
	eventQueue = &ScheduledEvents{}
	heap.Init(eventQueue)
	initialState()

	quit := make(chan bool)
	go schedule(eventQueue, quit)
	http.HandleFunc("/status", statusHandler)
	go http.ListenAndServe(":8081", nil)
	signalHandler(quit)
}
