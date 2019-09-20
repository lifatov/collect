package main

import (
	"log"
	"math/rand"
	"sync"
)

// problem1 implementation for provided task (see problem2.go for advanced version)
func problem1() {

	log.Printf("problem1: started --------------------------------------------")

	// Quit all go routines after
	// a total of exactly 100 random
	// numbers have been printed.
	const (
		MaxTotalNumbers    = 100 // maximum numbers to generate
		MaxTotalGenerators = 10  // goroutines for generators
	)

	// channel for notifications about generated numbers
	generated := make(chan interface{})

	// channel to say generators - "that's enough"
	done := make(chan interface{})

	wg := sync.WaitGroup{}

	// according to task - I can't change or edit signature of existing function:
	//   func printRandom1(slot int)
	// to prevent creation any global variables (for channels and for WaitGroup)
	// I'll use local closure with same name and signature
	var printRandom1 = func(slot int) {
		wg.Add(1)
		defer wg.Done()

		for inx := 0; inx < 25; inx++ {
			select {
			case <-done:
				return
			case generated <- struct{}{}:
				log.Printf("problem1: slot=%03d count=%05d rand=%f", slot, inx, rand.Float32())
			}
		}
	}

	for inx := 0; inx < MaxTotalGenerators; inx++ {

		go printRandom1(inx)

	}

	// We will count events (generated numbers)
	total := 0
	for range generated {
		total++

		if total >= MaxTotalNumbers {
			break
		}
	}

	// "That's enough, please stop."
	close(done)

	// Waiting goroutines to finish
	wg.Wait()

	log.Printf("problem1: finished --------------------------------------------")
}

// Not used
func printRandom1(slot int) {

	for inx := 0; inx < 25; inx++ {

		log.Printf("problem1: slot=%03d count=%05d rand=%f", slot, inx, rand.Float32())

	}
}
