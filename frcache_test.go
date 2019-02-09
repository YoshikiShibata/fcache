// Copyright © 2019 Yoshiki Shibata. All rights reserved.

package frcache

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSimpleGet(t *testing.T) {
	const (
		interval = 5 * time.Minute
		timeout  = 100 * time.Millisecond
	)
	msg := fmt.Sprintf("%x", time.Now().Unix())
	f := func() (interface{}, error) {
		return msg, nil
	}

	fc := New(f, interval, timeout)
	defer fc.Stop()

	result := fc.Get().(string)
	if result != msg {
		t.Errorf("result is %q, but want %q", result, msg)
	}
}

func TestNoResultAtTimeout(t *testing.T) {
	const (
		interval = 5 * time.Minute
		timeout  = 100 * time.Millisecond
	)
	msg := fmt.Sprintf("%x", time.Now().Unix())
	f := func() (interface{}, error) {
		time.Sleep(2 * timeout)
		return msg, nil
	}

	fc := New(f, interval, timeout)
	defer fc.Stop()

	result := fc.Get().(string)
	if result != msg {
		t.Errorf("result is %q, but want %q", result, msg)
	}
}

func TestNoResultAtTimeout2(t *testing.T) {
	const (
		interval = 5 * time.Minute
		timeout  = 100 * time.Millisecond
	)
	msg := fmt.Sprintf("%x", time.Now().Unix())

	count := 0
	f := func() (interface{}, error) {
		time.Sleep(2 * timeout)
		if count < 3 {
			count++
			return nil, errors.New("No result yet")
		}
		return msg, nil
	}

	fc := New(f, interval, timeout)
	defer fc.Stop()

	result := fc.Get().(string)
	if result != msg {
		t.Errorf("result is %q, but want %q", result, msg)
	}
}

func TestManyRequestes(t *testing.T) {
	const (
		interval = 5 * time.Minute
		timeout  = 100 * time.Millisecond
	)

	msg := fmt.Sprintf("%x", time.Now().Unix())
	var callCount uint64
	f := func() (interface{}, error) {
		time.Sleep(2 * timeout)
		atomic.AddUint64(&callCount, 1)
		return msg, nil
	}

	fc := New(f, interval, timeout)
	defer fc.Stop()

	const loopCount = 1000

	var wg sync.WaitGroup
	wg.Add(loopCount)
	for i := 0; i < loopCount; i++ {
		go func() {
			result := fc.Get().(string)
			if result != msg {
				t.Errorf("result is %q, but want %q", result, msg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	cc := atomic.AddUint64(&callCount, 0)
	if cc != 1 {
		t.Errorf("callCount is %d, but want 1", cc)
	}
}

func TestManyRequestes2(t *testing.T) {
	const (
		interval = 200 * time.Millisecond
		timeout  = 300 * time.Millisecond
	)

	msg := fmt.Sprintf("%x", time.Now().Unix())
	var callCount uint64
	f := func() (interface{}, error) {
		time.Sleep(2 * timeout)
		atomic.AddUint64(&callCount, 1)
		return msg, nil
	}

	fc := New(f, interval, timeout)
	defer fc.Stop()

	const loopCount = 1000

	var wg sync.WaitGroup
	wg.Add(loopCount)
	for i := 0; i < loopCount; i++ {
		go func() {
			result := fc.Get().(string)
			if result != msg {
				t.Errorf("result is %q, but want %q", result, msg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	cc := atomic.AddUint64(&callCount, 0)
	if cc != 1 {
		t.Errorf("callCount is %d, but want 1", cc)
	}
}
