package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

// problem2 implementation with pipelines and FanIn pattern to aggregate results from different goroutines
func problem2() {

	log.Printf("problem2: started --------------------------------------------")

	const (
		MaxTotalNumbers        = 100 // maximum numbers to generate
		MaxTotalGenerators     = 10  // goroutines for generators
		MaxNumbersPerGenerator = 25  // limit of generated numbers per generator
	)

	// channel to say generators - "that's enough"
	done := make(chan interface{})

	// initiate batch of generators
	generators := make([]<-chan interface{}, MaxTotalGenerators)

	for inx := 0; inx < MaxTotalGenerators; inx++ {
		//
		// each generator will be started in separate goroutine
		// pipeline provides ability to use different types of algorithms
		// to generate values of different types, examples:
		//
		//   randomFloatAlgorithm - generates random Float32 as result
		//   randomUUIDAlgorithm  - generates string with random UUID
		//
		generators[inx] = generator(done, inx, MaxNumbersPerGenerator, randomFloatAlgorithm)
	}

	total := 0
	for item := range fanIn(done, generators...) {
		total++

		if total > MaxTotalNumbers {
			break
		}

		log.Printf("problem2: slot=%03d count=%05d rand=%v",
			item.(Result).Slot, item.(Result).Index, item.(Result).Value)
	}

	// we should notify goroutines to stop processing
	close(done)

	log.Printf("problem2: finished --------------------------------------------")
}

// Result container to forward result with additional info about producer (if we need to have this in main thread)
type Result struct {
	Slot  int
	Index int
	Value interface{}
}

// randomFloatAlgorithm for generators
var randomFloatAlgorithm = func() interface{} {
	return rand.Float32()
}

// randomUUIDAlgorithm naive uuid for generators
var randomUUIDAlgorithm = func() interface{} {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, 12)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// generator returns channel and populates it with generated results
var generator = func(
	done <-chan interface{}, // channel to stop generator in case if we have enough numbers in common stream
	slot int,                // slot id to identify producer of value in common stream of results
	count int,               // count of values to generate
	fn func() interface{},   // function to generate single value
) <-chan interface{} {

	randomStream := make(chan interface{})

	go func() {
		defer close(randomStream)

		for inx := 0; inx < count; inx++ {
			item := Result{
				Slot:  slot,
				Index: inx,
				Value: fn(),
			}

			select {
			case <-done:
				return
			case randomStream <- item:
			}
		}
	}()

	return randomStream
}

// fanIn implements pattern FanIn to join results from multiple goroutines in one common stream (channel)
var fanIn = func(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup

	// merged output from 'channels'
	multiplexedStream := make(chan interface{})

	// extract data from channel to common output
	multiplex := func(c <-chan interface{}) {
		defer wg.Done()

		for i := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- i:
			}
		}
	}

	wg.Add(len(channels))

	// forward data from all channels to common output
	for _, c := range channels {
		go multiplex(c)
	}

	// Waiting for finish of merging data
	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	// return common channel to consume values from incoming channels
	return multiplexedStream
}
