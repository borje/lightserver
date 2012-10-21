package main

import (
	"bufio"
	"encoding/json"
	"github.com/cpucycle/astrotime"
	"log"
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
	action   Action
	weekdays string
	time     string
}

type ScheduledEvent struct {
	action Action
	time   time.Time
}

type ScheduledEvents []ScheduledEvent

func (se ScheduledEvents) Len() int           { return len(se) }
func (se ScheduledEvents) Swap(i, j int)      { se[i], se[j] = se[j], se[i] }
func (se ScheduledEvents) Less(i, j int) bool { return se[i].time.Before(se[j].time) }

func eventsForDay(now time.Time, schedule []ScheduleConfigItem) (events ScheduledEvents) {
	events = make(ScheduledEvents, 0, 7)
	weekDayToSelect := now.Weekday()
	for _, v := range schedule {
		weekdays := strings.Split(v.weekdays, ",")
		for _, wdStr := range weekdays {
			wd, _ := strconv.Atoi(wdStr)
			if weekDayToSelect == time.Weekday(wd) {
				var hour int
				var minute int
				if v.time == "SUNSET" {
					sunset := astrotime.CalcSunset(now, LATITUDE, LONGITUDE)
					hour = sunset.Hour()
					minute = sunset.Minute()
				} else if v.time == "SUNRISE" {
					sunset := astrotime.CalcSunrise(now, LATITUDE, LONGITUDE)
					hour = sunset.Hour()
					minute = sunset.Minute()
				} else if v.time[2] == ':' {
					timeStr := strings.Split(v.time, ":")
					hour, _ = strconv.Atoi(timeStr[0])
					minute, _ = strconv.Atoi(timeStr[1])
				}
				newPos := len(events)
				events = events[0 : newPos+1]
				events[newPos] = ScheduledEvent{v.action, time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())}
			}
		}
	}
	sort.Sort(events)
	return
}

func nextActionAfter(now time.Time, schedule []ScheduleConfigItem) (Action, time.Time) {
	for {
		for _, event := range eventsForDay(now, schedule) {
			if event.time.After(now) {
				return event.action, event.time
			}
		}
		nextDay := now.Add(OneDay)
		now = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())
	}
	log.Fatal("Should not return here")
	return TurnOn, now
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
		log.Println("Error executing Tellstick action: ", cmd)
		// Some kind error handling
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

func schedule(configuration []ScheduleConfigItem, quit chan bool) {
	for {
		now := time.Now()
		action, nextTime := nextActionAfter(now, configuration)
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

func signalHandler(quit chan bool) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChannel
	quit <- true
	log.Println("Received signal: ", sig)
}

func main() {
	configuration := []ScheduleConfigItem{
		{TurnOn, "1,2,3,4,5,6,0", "SUNSET"},
		{TurnOff, "0,1,2,3,4", "22:15"},
		{TurnOff, "5,6", "23:00"},
		{TurnOn, "1,2,3,4,5", "06:45"},
		{TurnOff, "1,2,3,4,5", "SUNRISE"},
	}
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
	action, _ := nextActionAfter(now, configuration)
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
	/*time.Sleep(time.Second / 2)*/
}
