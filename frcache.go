// Copyright © 2019 Yoshiki Shibata. All rights reserved.

package frcache

import (
	"log"
	"time"
)

// Func defines a function to be executed to cache its result.
// This function will be execued periodically to cache its result.
type Func func() (interface{}, error)

type FRCache struct {
	f        Func
	interval time.Duration
	timeout  time.Duration

	cachedResult     interface{}
	getRequestChan   chan chan interface{}
	getResponseChans []chan interface{}

	executing bool

	intervalTicker *time.Ticker
	timeoutChan    <-chan time.Time

	resultChan chan interface{}
	stopChan   chan struct{}
}

func New(f Func, interval, timeout time.Duration) *FRCache {
	fc := &FRCache{
		f:                f,
		interval:         interval,
		timeout:          timeout,
		cachedResult:     nil,
		getRequestChan:   make(chan chan interface{}),
		getResponseChans: nil,
		executing:        false,
		intervalTicker:   time.NewTicker(interval),
		timeoutChan:      nil,
		resultChan:       make(chan interface{}),
		stopChan:         make(chan struct{}),
	}

	go fc.monitor()
	return fc
}

// Get invokes Func and waits for the timeout period, and then return
// any cached value.
func (fc *FRCache) Get() interface{} {
	getChan := make(chan interface{})
	fc.getRequestChan <- getChan
	return <-getChan
}

func (fc *FRCache) Stop() {
	close(fc.getRequestChan)
	close(fc.stopChan)
}

func (fc *FRCache) monitor() {
	for {
		select {
		case responseChan := <-fc.getRequestChan:
			fc.getResponseChans = append(fc.getResponseChans, responseChan)
			fc.executeFunction()
			if fc.timeoutChan == nil {
				fc.timeoutChan = time.After(fc.timeout)
			}
		case <-fc.intervalTicker.C:
			fc.executeFunction()
		case result := <-fc.resultChan:
			fc.executing = false
			fc.cachedResult = result
			fc.returnCachedResult()
		case <-fc.timeoutChan:
			fc.timeoutChan = nil
			fc.returnCachedResult()
		case <-fc.stopChan:
			fc.intervalTicker.Stop()
			return
		}
	}
}

func (fc *FRCache) execute() {
	for {
		result, err := fc.f()
		if err != nil {
			log.Printf("f() failed: %v", err)
			continue
		}

		fc.resultChan <- result
		return
	}
}

func (fc *FRCache) executeFunction() {
	if fc.executing {
		return
	}
	fc.executing = true
	go fc.execute()
}

func (fc *FRCache) returnCachedResult() {
	// For the first time, the cachedResult may be nil, then do nothing.
	if fc.cachedResult == nil {
		return
	}

	// For optimization, check the number of "Get" requests.
	if len(fc.getResponseChans) == 0 {
		return
	}

	// Send the cachedResult to all requesters.
	reqs := fc.getResponseChans
	cache := fc.cachedResult

	fc.getResponseChans = nil
	go func() {
		for _, req := range reqs {
			req <- cache
		}
	}()
}
