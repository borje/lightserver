package main

import (
	"bufio"
	"lightserver/scheduler"
	"net/http"
	"os"
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
	writer := bufio.NewWriter(w)
	writer.ReadFrom(file)
	writer.Flush()
}
