package pubsub

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type Agent[T any] struct {
	mu         sync.Mutex
	subs       map[string][]chan T
	quit       chan struct{}
	closed     bool
	logDir     string
	topicQueue map[string][]T
}

func NewAgent[T any](logDir string) *Agent[T] {
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chmod(logDir, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return &Agent[T]{
		subs:       make(map[string][]chan T),
		quit:       make(chan struct{}),
		logDir:     logDir,
		topicQueue: make(map[string][]T),
	}
}

func (b *Agent[T]) Publish(topic string, msg T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	logFilePath := b.logDir + "/" + topic + ".txt"
	b.topicQueue[topic] = append(b.topicQueue[topic], msg)
	if err := b.writeToFile(logFilePath, msg); err != nil {
		log.Println(err)
	}
	for _, ch := range b.subs[topic] {
		ch <- msg
	}
}

func (b *Agent[T]) Subscribe(topic string) <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	ch := make(chan T)
	b.subs[topic] = append(b.subs[topic], ch)
	return ch
}

func (b *Agent[T]) writeToFile(filePath string, msg T) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(fmt.Sprintf("%v\n", msg)); err != nil {
		return err
	}
	return nil
}

func (b *Agent[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	close(b.quit)
	for _, ch := range b.subs {
		for _, sub := range ch {
			close(sub)
		}
	}
	b.saveTopicQueues()
}

func (b *Agent[T]) saveTopicQueues() {
	for topic, queue := range b.topicQueue {
		filePath := b.logDir + "/" + topic + ".txt"
		_, err := os.Stat(filePath)
		fileExists := !os.IsNotExist(err)

		for _, msg := range queue {
			if !fileExists {
				if err := b.writeToFile(filePath, msg); err != nil {
					log.Println(err)
				}
			}
		}
	}
}
