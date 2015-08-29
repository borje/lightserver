package main

import (
	"io"
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

func fileReturnHandler(filename string) func(http.ResponseWriter, *http.Request) {
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
