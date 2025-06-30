package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/russross/blackfriday/v2"
)

// From blackfriday docs: as simple as getting input into a byteslice
// and then calling blackfriday.Run(input) <- we will get the "input" path
// from fsnotify, I think that is the way to go about it.

type Broker struct {
	id          int
	rwm         sync.RWMutex
	subscribers map[int]chan []byte
	fp          chan string
}

func NewBroker() *Broker {
	b := &Broker{
		id:          0,
		rwm:         sync.RWMutex{},
		subscribers: make(map[int]chan []byte),
		fp:          make(chan string),
	}

	return b
}

func (b *Broker) AddSubscriber() (int, chan []byte) {
	b.rwm.Lock()
	defer b.rwm.Unlock()
	b.id += 1
	b.subscribers[b.id] = make(chan []byte)
	return b.id, b.subscribers[b.id]
}

func (b *Broker) RemoveSubscriber(id int) {
	b.rwm.Lock()
	defer b.rwm.Unlock()
	ch, ok := b.subscribers[id]
	if ok {
		close(ch)
		delete(b.subscribers, id)
	}
}

func (b *Broker) Broadcast(fp string) {
	b.rwm.RLock()
	defer b.rwm.RUnlock()

	var html []byte
	fileByteContent, err := os.ReadFile(fp)

	if err != nil {
		errorMsg := fmt.Sprintf("failed to read file: %v", err)
		html = []byte(errorMsg)
	} else {
		html = blackfriday.Run(fileByteContent)
	}

	for _, ch := range b.subscribers {
		select {
		case ch <- html:
		default:
		}
	}
}

func (b *Broker) Run() {
	for fp := range b.fp {
		b.Broadcast(fp)
	}
}
