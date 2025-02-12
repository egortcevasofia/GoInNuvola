package core

import (
	"errors"
	"log"
	"sync"
)

const (
	_                     = iota
	EventDelete EventType = iota
	EventPut              = iota
)

type TransactionLogger interface { //порт в гексагональной архитектуре
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error
	ReadEvents() (<-chan Event, <-chan error)
	Run()
}

type Frontend interface { //порт в гексагональной архитектуре
	Start(kv *KeyValueStore) error
}

type Event struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}
type EventType byte

type KeyValueStore struct {
	M        StoreWithMutex
	Transact TransactionLogger
}

func NewKeValueStore(tl TransactionLogger) *KeyValueStore {
	return &KeyValueStore{
		M:        StoreWithMutex{m: make(map[string]string)},
		Transact: tl,
	}
}

type StoreWithMutex struct {
	mytex sync.RWMutex
	m     map[string]string
}

var ErrorNoSuchKey = errors.New("no such key")

func (store *KeyValueStore) Put(key string, value string) error {
	store.M.mytex.Lock()
	store.M.m[key] = value
	store.M.mytex.Unlock()
	return nil
}

func (store *KeyValueStore) Get(key string) (string, error) {
	store.M.mytex.RLock()
	value, ok := store.M.m[key]
	store.M.mytex.RUnlock()
	if !ok {
		return "", ErrorNoSuchKey
	}
	return value, nil
}

func (store *KeyValueStore) Delete(key string) error {
	store.M.mytex.Lock()
	delete(store.M.m, key)
	store.M.mytex.Unlock()
	return nil
}

func (store *KeyValueStore) Restore() error {
	var err error

	events, errors := store.Transact.ReadEvents()
	count, ok, e := 0, true, Event{}

	for ok && err == nil {
		select {
		case err, ok = <-errors:

		case e, ok = <-events:
			switch e.EventType {
			case EventDelete: // Got a DELETE event!
				err = store.Delete(e.Key)
				count++
			case EventPut: // Got a PUT event!
				err = store.Put(e.Key, e.Value)
				count++
			}
		}
	}

	log.Printf("%d events replayed\n", count)

	store.Transact.Run()

	go func() {
		for err := range store.Transact.Err() {
			log.Print(err)
		}
	}()

	return err
}
