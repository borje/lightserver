package scheduler

import (
	"container/heap"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cpucycle/astrotime"
)

const (
	LATITUDE  = 58.410807
	LONGITUDE = -15.6213727
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
}

type Scheduler struct {
	configFile  string
	eventQueue  *ScheduledEvents
	configItems []ScheduleConfigItem
}

type ScheduledEvent struct {
	Device int
	Action Action
	Time   time.Time
}

type ScheduleConfigItem struct {
	Device   int
	Weekdays string
	TimeFrom string
	TimeTo   string
}

type ScheduledEvents []ScheduledEvent

func NewSchedulerFromReader(r io.Reader) *Scheduler {
	scheduler := &Scheduler{
		configFile: "",
		eventQueue: &ScheduledEvents{},
	}
	jsonDecoder := json.NewDecoder(r)
	err := jsonDecoder.Decode(&scheduler.configItems)
	if err != nil {
		log.Fatal(err)
	}
	heap.Init(scheduler.eventQueue)
	return scheduler
}

func NewScheduler(configFile string) *Scheduler {
	scheduler := &Scheduler{
		configFile: configFile,
		eventQueue: &ScheduledEvents{},
	}
	scheduler.configItems = getConfiguration(configFile)
	heap.Init(scheduler.eventQueue)
	scheduler.initialState()
	return scheduler
}

func (s *Scheduler) EventQueue() ScheduledEvents {
	return *s.eventQueue
}

// Helper functions
func timeFromString(theDay time.Time, clock string) time.Time {
	timeStr := strings.Split(clock, ":")
	hour, _ := strconv.Atoi(timeStr[0])
	minute, _ := strconv.Atoi(timeStr[1])
	return time.Date(theDay.Year(), theDay.Month(), theDay.Day(), hour, minute, 0, 0, theDay.Location())
}

/* Functions for sorting ScheduledEvents */
func (se ScheduledEvents) Len() int           { return len(se) }
func (se ScheduledEvents) Swap(i, j int)      { se[i], se[j] = se[j], se[i] }
func (se ScheduledEvents) Less(i, j int) bool { return !se[i].Time.After(se[j].Time) }

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
				if onEvent.Time.Before(offEvent.Time) {
					events = append(events, onEvent)
					events = append(events, offEvent)
				}
			}
		}
	}
	return
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

func getConfiguration(file string) []ScheduleConfigItem {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal("Error opening ", file, " Error: ", err)
	}

	jsonDecoder := json.NewDecoder(f)
	var config []ScheduleConfigItem
	err = jsonDecoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func (this *Scheduler) AddEventsForDay(day time.Time) {
	for _, event := range eventsForDay(day, this.configItems) {
		heap.Push(this.eventQueue, event)
	}
}

func addEventForDay(eq *ScheduledEvents, configItems []ScheduleConfigItem, day time.Time) {
	for _, event := range eventsForDay(day, configItems) {
		log.Println("Adding ", event.Action, " event @", event.Time)
		heap.Push(eq, event)
	}
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

func (this *Scheduler) initialState() time.Time {
	log.Println("Initial states")
	now := time.Now()
	currentDay := now
	for _, device := range configuredDevices(this.configItems) {
		actionFound := false
		for !actionFound {
			for i := range *this.eventQueue {
				event := (*this.eventQueue)[this.eventQueue.Len()-1-i] // Iterate the heap backwards
				if event.Device == device && event.Time.Before(now) {
					doTellstickAction(device, event.Action)
					actionFound = true
					break
				}
			}
			if !actionFound {
				addEventForDay(this.eventQueue, this.configItems, currentDay)
				currentDay = currentDay.AddDate(0, 0, -1)
			}
		}
	}
	// remove old events
	for this.eventQueue.Len() > 0 {
		next := heap.Pop(this.eventQueue).(ScheduledEvent)
		log.Printf("Removing %d @ %s", next.Device, next.Time)
		if next.Time.After(now) {
			heap.Push(this.eventQueue, next)
			break
		}
	}
	return now
}

func (this *Scheduler) Schedule(quit chan bool) {
	currentDay := time.Now()
	for {
		for this.eventQueue.Len() > 0 {
			event := (*this.eventQueue)[0]
			log.Printf("Next event: %s @ %s (device %d)", event.Action, event.Time, event.Device)
			if now := time.Now(); now.Before(event.Time) {
				log.Printf("Sleeping for %s", event.Time.Sub(now))
				timer := time.NewTimer(event.Time.Sub(now))
				select {
				case <-quit:
					log.Println("Quit scheduling")
					quit <- true
					return
				case <-timer.C:
					doTellstickAction(event.Device, event.Action)
				}
			} else {
				doTellstickAction(event.Device, event.Action)
			}
			heap.Pop(this.eventQueue)
		}
		currentDay = currentDay.AddDate(0, 0, 1)
		addEventForDay(this.eventQueue, this.configItems, currentDay)
	}
}
