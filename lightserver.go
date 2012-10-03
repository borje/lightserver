package main

import (
	"bufio"
	"fmt"
	"github.com/cpucycle/astrotime"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	OneDay = time.Hour * 24
	LATITUDE           = 58.410807
	LONGITUDE          = -15.6213727
	turnOnBeforeSunset = -time.Hour
	NIGHT_OFF_HOUR     = 22
	NIGHT_OFF_MINUTE   = 0
	/*NIGHT_OFF_HOUR     = 22*/
	/*NIGHT_OFF_MINUTE   = 05*/
)

type Action int
const (
	TurnOn Action = iota
	TurnOff
)

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
		t = nightOffTime.Add(OneDay)
	}
	return
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("tdtool", "--list")
	stdout, _ := cmd.StdoutPipe()
	reader := bufio.NewReader(stdout)
	cmd.Start()
	str, _ := reader.ReadString('\n')
	nrReceivers := strings.Split(str, ":")[1]
	nRcv, _ := strconv.ParseInt(nrReceivers, 0, 32)
	fmt.Print(nRcv)
	fmt.Fprintf(w, "status: %s\n", str)
	fmt.Fprintf(w, "Receivers: '%s'\n", nrReceivers)
	fmt.Fprintf(w, "Receivers: %d\n", nRcv)
}

func main() {
	now := time.Now()
	nightOnTime := astrotime.CalcSunset(now, LATITUDE, LONGITUDE).Add(turnOnBeforeSunset)
	nightOffTime := time.Date(now.Year(), now.Month(), now.Day(), NIGHT_OFF_HOUR, NIGHT_OFF_MINUTE, 0, 0, now.Location())
	action, nextTime := nextActionTime(now)
	fmt.Println(action, nextTime)
	/*var timer *time.Timer*/
	if nightOnTime.Before(now) && now.Before(nightOffTime) {
		fmt.Println("TÄND")
		fmt.Printf("Kommer att släcka kl. %2d:%02d\n", nextTime.Hour(), nextTime.Minute())
		/*dur := nightOffTime.Sub(time.Now())*/
		/*timer = time.NewTimer(dur)*/
	} else {
		fmt.Println("SLÄCK")
		fmt.Printf("Kommer att tända kl. %2d:%02d\n", nextTime.Hour(), nextTime.Minute())
	}
	/*<-timer.C*/
	return
	http.HandleFunc("/status", statusHandler)
	http.ListenAndServe(":8081", nil)
}
