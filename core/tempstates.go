package core

import (
	"fmt"
	"time"
)

const TICK_SECONDS = 2

/**
 * this is a simple in-memory cache structure where entries must be given a fixed time to live
 */

type TemporaryPropertyStore struct {
	values map[string]*TemporaryProperty
}

type TemporaryProperty struct {
	value     string
	starttime int64
	ttl       int64
}

var Stores map[string]*TemporaryPropertyStore

func NewTemporaryPropertyStore(name string) *TemporaryPropertyStore {

	_, exists := Stores[name]
	if exists {
		fmt.Printf("TemporaryPropertyStore %s already exists\n", name)
		return nil
	}

	store := new(TemporaryPropertyStore)
	store.values = make(map[string]*TemporaryProperty)
	go store.monitor()
	return store

}

func GetTemporaryPropertyStore(name string) *TemporaryPropertyStore {
	s, exists := Stores[name]
	if !exists {
		return nil
	}

	return s
}

func (this *TemporaryPropertyStore) GetEntry(key string) *string {

	val, exists := this.values[key]
	if !exists {
		return nil
	}
	return &(val.value)
}

func (this *TemporaryPropertyStore) AddEntry(key string, val string, ttlMillis int) {

	now := time.Now().UnixNano()
	entry := &TemporaryProperty{key, now, int64(ttlMillis) * int64(time.Millisecond)}
	this.values[key] = entry

	delay := time.Duration(int64(ttlMillis) * int64(time.Millisecond))

	timeoutCheck := func() {
		this.clear()
	}
	time.AfterFunc(delay, timeoutCheck)
}

func (this *TemporaryPropertyStore) RemoveEntry(key string) {
	_, exists := this.values[key]
	if exists {
		delete(this.values, key)
	}
}

func (this *TemporaryPropertyStore) clear() {
	now := time.Now().UnixNano()
	for k, v := range this.values {
		if now-v.starttime >= v.ttl {
			fmt.Printf("Deleting ephemeral entry\n")
			delete(this.values, k)
		} else {
			fmt.Printf("Leaving ephemeral entry\n")
		}
	}
}

func (this *TemporaryPropertyStore) monitor() {

	ticker := time.Tick(TICK_SECONDS * time.Second)
	//	ticker := time.NewTicker(time.Duration(int64(time.Second) * TICK_SECONDS))
	//	defer ticker.Stop()

	for now := range ticker {
		_ = now
		this.clear()
	}
}
