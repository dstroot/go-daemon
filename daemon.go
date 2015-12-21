package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	start := time.Now()
	log.Println("Starting application...")

	// When SIGINT or SIGTERM is caught write to the sigs channel
	sigs := make(chan os.Signal)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Catch signal and send done
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	closer := make(chan struct{})
	wg := new(sync.WaitGroup)

	// Create a goroutine that does imaginary work
	wg.Add(1)
	go func() {
		log.Println("Starting work goroutine...")
		defer wg.Done()

		for {
			// Listen for close message
			select {
			case _ = <-closer:
				return

			default:
			}

			// Do some hard work here!
		}
	}()

	// Wait for done message
	<-done
	close(closer)
	log.Println("Received quit. Sending shutdown and waiting on goroutines...")

	// Block until wait group counter gets to zero (no more goroutines)
	wg.Wait()
	log.Printf("Running for: %s\n", time.Since(start))
	log.Println("Goodbye!")
}
