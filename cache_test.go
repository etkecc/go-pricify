package pricify

import (
	"testing"
)

func newMockData() *Data {
	return &Data{}
}

func TestSetAndGetCache(t *testing.T) {
	cached = nil // Reset cache before test
	data := newMockData()
	setCache(data)
	got := getCache()
	if got != data {
		t.Errorf("expected %v, got %v", data, got)
	}
}

func TestGetCacheNil(t *testing.T) {
	cached = nil // Reset cache before test
	got := getCache()
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestCacheOverwrite(t *testing.T) {
	cached = nil // Reset cache before test
	data1 := newMockData()
	data2 := newMockData()
	setCache(data1)
	setCache(data2)
	got := getCache()
	if got != data2 {
		t.Errorf("expected %v, got %v", data2, got)
	}
}

func TestConcurrentSetAndGet(t *testing.T) {
	cached = nil // Reset cache before test
	data1 := newMockData()
	data2 := newMockData()
	done := make(chan bool)

	go func() {
		setCache(data1)
		done <- true
	}()
	go func() {
		setCache(data2)
		done <- true
	}()
	<-done
	<-done

	got := getCache()
	if got != data1 && got != data2 {
		t.Errorf("expected %v or %v, got %v", data1, data2, got)
	}
}
