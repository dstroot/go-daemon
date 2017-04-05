package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

// Metrics holds our metrics for reporting
type Metrics struct {
	Program string
	Version string
	RunTime string
	Widgets int
}

var report Metrics           // Global Metrics
var start = time.Now().UTC() // Global Start

// status returns a JSON object with the current
// metrics, counts, etc. of the running program
func status(w http.ResponseWriter, req *http.Request) {
	report.RunTime = fmt.Sprintf("%v", time.Since(start))
	js, err := json.Marshal(report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type (not a loopback)
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// setupEnvironment parses our config file (using Viper)
func configure() {
	viper.AddConfigPath("./config/")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func housekeeping() {
	configure()
	address := GetLocalIP()
	path := strings.Split(os.Args[0], "/")
	report.Program = path[len(path)-1]
	report.Version = "0.0.1"
	log.Printf("%s running. Get status: 'http://%s%s/status'", report.Program, address, viper.GetString("port"))
}

func main() {
	housekeeping()

	// Report metrics (JSON via http) via separate goroutine
	go func() {
		http.Handle("/status", http.HandlerFunc(status))
		err := http.ListenAndServe(viper.GetString("port"), nil)
		if err != nil {
			log.Fatal("ListenAndServe error: ", err)
		}
	}()

	// When SIGINT or SIGTERM is caught write to the sigs channel
	// buffer to make sure we catch it
	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Catch signal and send done via separate goroutine
	go func() {
		sig := <-sigs
		if sig == os.Interrupt {
			fmt.Println()
			done <- true
		}
	}()

	closer := make(chan struct{})
	wg := new(sync.WaitGroup)

	// Create main goroutine that does imaginary work
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Listen for close message
		for {
			select {
			case _ = <-closer:
				return

			default:
			}

			// Do some hard work here!
			report.Widgets++
			log.Printf("another %s...", viper.GetString("widget"))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Wait for done message
	<-done
	// Send close to goroutine
	close(closer)
	// Block until wait group counter gets to zero (main goroutine has stopped)
	wg.Wait()
	log.Printf("%s ran for: %s\n", report.Program, time.Since(start))
	log.Printf("%d %ss produced", report.Widgets, viper.GetString("widget"))
	log.Println("goodbye!")
}
