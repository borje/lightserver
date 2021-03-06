package main

import (
	"io"
	"lightserver/scheduler"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func StatusWrapper(s *scheduler.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		eventsCopy := s.EventQueue()
		sort.Sort(eventsCopy)
		rend.JSON(w, http.StatusOK, eventsCopy)
	}
}

type buildInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
}

func infoHandler(w http.ResponseWriter, req *http.Request) {
	info := buildInfo{Version: currentVersion(),
		BuildTime: buildTime()}
	rend.JSON(w, http.StatusOK, info)
}

func fileReturnHandler(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		file, err := os.Open(filename)
		if err != nil {
			http.Error(w, "Can't open file", http.StatusInternalServerError)
			return
		}
		defer func() {
			file.Close()
		}()

		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusInternalServerError)
			return
		}
	}
}

func scheduleHandler(w http.ResponseWriter, req *http.Request) {
	f, _ := os.Open(*configFile)
	s := scheduler.NewSchedulerFromReader(f)
	f.Close()

	start := req.URL.Query().Get("start")

	startDate, err := time.ParseInLocation("2006-01-02", start, time.Local)
	if err != nil {
		log.Println(err)
		http.Error(w, "unknown start date", http.StatusInternalServerError)
		return
	}

	end := req.URL.Query().Get("end")
	endDate, err := time.ParseInLocation("2006-01-02", end, time.Local)
	if err != nil {
		log.Println(err)
		http.Error(w, "unknown end date", http.StatusInternalServerError)
		return
	}

	var periods []scheduler.OnPeriod
	for d := startDate; !d.After(endDate); d = d.Add(time.Hour * 24) {
		periods = append(periods, s.Periods(d)...)
	}
	rend.JSON(w, http.StatusOK, periods)
}

func controlHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	devstr := vars["device"]
	action := vars["action"]
	if device, err := strconv.Atoi(devstr); err == nil {
		if strings.ToLower(action) == "on" {
			go scheduler.DoTellstickAction(device, scheduler.TurnOn)
		} else {
			go scheduler.DoTellstickAction(device, scheduler.TurnOff)
		}
	}

}
