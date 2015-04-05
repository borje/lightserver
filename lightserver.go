package main

import (
	"flag"
	"lightserver/scheduler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/unrolled/render"
)

//go:generate /bin/sh ./generate_build_info.sh

var configFile = flag.String("configfile", "config.json", "The Config")
var debug = flag.Bool("debug", false, "Print logging to stderr instead of file")

const (
	LOG_FILE = "lightserver.log"
)

func signalHandler(quit chan bool) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChannel
	quit <- true
	<-quit
	log.Println("Received signal: ", sig)
}

var rend *render.Render

func main() {
	flag.Parse()
	defer func() {
		log.Println("Exiting")
	}()
	if !*debug {
		logfile, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(logfile)
			defer func() {
				logfile.Close()
			}()
		}
	}
	log.Println("Starting lightserver version ", currentVersion())
	rend = render.New(render.Options{IndentJSON: true})
	scheduler := scheduler.NewScheduler(*configFile)

	quit := make(chan bool)
	go scheduler.Schedule(quit)
	http.HandleFunc("/status", StatusWrapper(scheduler))
	http.HandleFunc("/info", infoHandler)
	go http.ListenAndServe(":8081", nil)
	signalHandler(quit)
}

func StatusWrapper(s *scheduler.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Println("/info is called")
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
