package main

import (
	"flag"
	"lightserver/scheduler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

import _ "net/http/pprof"

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

func logHandlerFunc(f func(w http.ResponseWriter, req *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		f(w, req)
		log.Printf("%s\t%-15s\t%s\tfrom %s", req.Method, req.RequestURI, time.Since(start), req.RemoteAddr)
	}
}

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logHandlerFunc(h.ServeHTTP)(w, req)
	})
}

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
	} else {
		log.SetFlags(log.Flags() | log.Lshortfile)
	}
	log.Println("Starting lightserver version ", currentVersion())
	rend = render.New(render.Options{IndentJSON: true})
	scheduler := scheduler.NewScheduler(*configFile)

	quit := make(chan bool)
	go scheduler.Schedule(quit)
	router := mux.NewRouter()
	router.HandleFunc("/status", logHandlerFunc(StatusWrapper(scheduler)))
	router.HandleFunc("/info", logHandlerFunc(infoHandler))
	router.HandleFunc("/config", logHandlerFunc(fileReturnHandler(*configFile)))
	router.HandleFunc("/log", logHandlerFunc(fileReturnHandler(LOG_FILE)))
	router.HandleFunc("/schedule/{year}/{month}/{day}", logHandlerFunc(scheduleHandler))
	router.PathPrefix("/").Handler(logHandler(http.FileServer(http.Dir("static"))))
	http.Handle("/", router)
	go http.ListenAndServe(":8081", nil)
	signalHandler(quit)
}
