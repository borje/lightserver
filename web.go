package main

import (
	"io"
	"lightserver/scheduler"
	"net/http"
	"os"
	"time"
)

func StatusWrapper(s *scheduler.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		rend.JSON(w, http.StatusOK, s.EventQueue())
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

func logHandler(w http.ResponseWriter, req *http.Request) {
	file, err := os.Open(LOG_FILE)
	if err != nil {
		http.Error(w, "Can't open log file", http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Unable to write log", http.StatusInternalServerError)
		return
	}
}

func scheduleHandler(w http.ResponseWriter, req *http.Request) {
		f, _ := os.Open(*configFile)
		s := scheduler.NewSchedulerFromReader(f)
		s.AddEventsForDay(time.Now())
		defer f.Close()
		rend.JSON(w, http.StatusOK, s.EventQueue())
	}

