package pubsub

import (
	"log"
	"os"
	"sync"
)

type Agent struct {
	Mu      sync.Mutex
	Subs    map[string][]chan string
	Quit    chan struct{}
	Closed  bool
	LogFile *os.File
}

func NewAgent() *Agent {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return &Agent{
		Subs:    make(map[string][]chan string),
		Quit:    make(chan struct{}),
		LogFile: file,
	}
}

func (b *Agent) Publish(topic string, msg string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if b.Closed {
		return
	}
	if _, err := b.LogFile.WriteString(msg + "\n"); err != nil {
		log.Println(err)
	}
	for _, ch := range b.Subs[topic] {
		ch <- msg
	}
}

func (b *Agent) Subscribe(topic string) <-chan string {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if b.Closed {
		return nil
	}
	ch := make(chan string)
	b.Subs[topic] = append(b.Subs[topic], ch)
	return ch
}

func (b *Agent) Close() {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if b.Closed {
		return
	}
	b.Closed = true
	close(b.Quit)
	for _, ch := range b.Subs {
		for _, sub := range ch {
			close(sub)
		}
	}
	if err := b.LogFile.Close(); err != nil {
		log.Println(err)
	}
}
