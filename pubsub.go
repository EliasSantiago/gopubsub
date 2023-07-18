package pubsub

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type Valor interface {
	int64 | float64 | string
}

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

func (a *Agent[T]) Publish(topic string, msg T) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	logFilePath := a.logDir + "/" + topic + ".txt"
	a.topicQueue[topic] = append(a.topicQueue[topic], msg)
	if err := a.writeToFile(logFilePath, msg); err != nil {
		log.Println(err)
	}
	for _, ch := range a.subs[topic] {
		ch <- msg
	}
}

func (a *Agent[T]) Subscribe(topic string) <-chan T {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return nil
	}
	ch := make(chan T)
	a.subs[topic] = append(a.subs[topic], ch)
	return ch
}

func (a *Agent[T]) writeToFile(filePath string, msg T) error {
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

func (a *Agent[T]) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	a.closed = true
	close(a.quit)
	for _, ch := range a.subs {
		for _, sub := range ch {
			close(sub)
		}
	}
	a.saveTopicQueues()
}

func (a *Agent[T]) saveTopicQueues() {
	for topic, queue := range a.topicQueue {
		filePath := a.logDir + "/" + topic + ".txt"
		_, err := os.Stat(filePath)
		fileExists := !os.IsNotExist(err)

		for _, msg := range queue {
			if !fileExists {
				if err := a.writeToFile(filePath, msg); err != nil {
					log.Println(err)
				}
			}
		}
	}
}
