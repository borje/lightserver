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

/* Functions for sorting ScheduledEvents */
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
	const MAX_EVENTS_PER_DAY_PER_DEVICE = 8
	events = ScheduledEvents{}
	currentWeekDay := now.Weekday()
	for _, v := range schedule { // unused return value ?
		device := v.device

		weekdays := strings.Split(v.weekdays, ",")
		for _, dayInWeekString := range weekdays {
			dayInWeek, _ := strconv.Atoi(dayInWeekString)
			if currentWeekDay == time.Weekday(dayInWeek) {
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
				newEvent := ScheduledEvent{device, TurnOn, time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())}
				events = append(events, newEvent)

				// TURN OFF
				if v.timeTo[2] == ':'{
					timeStr := strings.Split(v.timeTo, ":")
					hour, _ = strconv.Atoi(timeStr[0])
					minute, _ = strconv.Atoi(timeStr[1])
				}
				newEvent = ScheduledEvent{device, TurnOff, time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())}
				events = append(events, newEvent)
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

func doTellstickAction(device int, action Action) {
	var tellstickCmd string
	if action == TurnOn {
		tellstickCmd = "--on"
	} else {
		tellstickCmd = "--off"
	}
	/*cmd := exec.Command("tdtool", tellstickCmd, "2")*/
	cmd := exec.Command("echo", tellstickCmd, strconv.Itoa(device))
	b, err := cmd.CombinedOutput()
	log.Println("Turning device", device, ":", action)
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
		device, action, next := nextActionAfter(now, getConfiguration())
		fmt.Println(device, next, action)
		now = next
	}
}

func schedule(configuration []ScheduleConfigItem, quit chan bool) {
	for {
		now := time.Now()
		device, action, nextTime := nextActionAfter(now, configuration)
		log.Printf("Next event: %s @ %s (device %d)", action, nextTime, device)
		untilNextAction := nextTime.Sub(now)
		timer := time.NewTimer(untilNextAction)
		select {
		case <-quit:
			log.Println("Quit scheduling")
			return
		case <-timer.C:
			log.Println("Executing action: ", action)
			doTellstickAction(device, action)
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

func getConfiguration() []ScheduleConfigItem {
	return []ScheduleConfigItem{
		/*{2, "1,2,3,4,5,6,0", "15:00", "22:15"},*/
		/*{2, "1,2,3,4,5,6,0", "07:15", "09:30"},*/
		/*{1, "1,2,3,4,5,6,0", "05:35", "11:00"},*/
		/*{1, "1,2,3,4,5,6,0", "13:00", "22:15"},*/
		{3, "1,2,3,4,5,6,0", "22:40", "22:41"},
		{4, "1,2,3,4,5,6,0", "22:40", "22:41"},
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

func main() {
	configuration := getConfiguration()
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

	// DEBUG
	for i := 0; i < 8; i = i + 1 {
		device, action, next := nextActionAfter(now, configuration)
		fmt.Println("DEBUG: ", device, next, action)
		now = next
	}

	/// Set correct light status at startup
	// TODO: incorrent assumption. Should iterate backwards in time instead
	for _, device := range configuredDevices(getConfiguration()) {
		now := time.Now()
		var nextAction Action
		for {
			nextDevice, action, next := nextActionAfter(now, configuration)
			if nextDevice == device {
				nextAction = action
				break
			}
			now = next
		}
		fmt.Println("Turning device inverse of ", device, " ", nextAction)
		if nextAction == TurnOff {
			go doTellstickAction(device, TurnOn)
		} else {
			go doTellstickAction(device, TurnOff)
		}
	}
	quit := make(chan bool)
	go schedule(configuration, quit)
	http.HandleFunc("/status", statusHandler)
	go http.ListenAndServe(":8081", nil)
	signalHandler(quit)
}
