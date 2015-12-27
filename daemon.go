package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Metrics holds our metrics for reporting
type Metrics struct {
	Name    string
	Version string
	Widgets int
}

var report Metrics

// status returns a JSON object with the current
// metrics, counts, etc. of the running program
func status(w http.ResponseWriter, req *http.Request) {
	js, err := json.Marshal(report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	path := strings.Split(os.Args[0], "/")
	report.Name = path[len(path)-1]
	report.Version = "0.0.1"
	log.Printf("%s running. Check status at 'http://localhost:8080/status'", report.Name)
	start := time.Now()

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

	// Report metrics (JSON via http) via separate goroutine
	go func() {
		http.Handle("/status", http.HandlerFunc(status))
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("ListenAndServe error: ", err)
		}
	}()

	// Create main goroutine that does imaginary work
	wg.Add(1)
	go func() {
		defer wg.Done()
		// log.Println("Starting main work goroutine...")

		// Listen for close message
		for {
			select {
			case _ = <-closer:
				return

			default:
			}

			// Do some hard work here!
			report.Widgets++
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Wait for done message
	<-done
	// Send close to goroutine
	close(closer)
	// Block until wait group counter gets to zero (main goroutine has stopped)
	wg.Wait()
	log.Printf("%s ran for: %s\n", report.Name, time.Since(start))
	log.Println("goodbye!")
}
