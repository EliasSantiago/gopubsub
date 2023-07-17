package pubsub

import (
	"log"
	"os"
	"sync"
)

type Agent struct {
	mu         sync.Mutex
	subs       map[string][]chan string
	quit       chan struct{}
	closed     bool
	logDir     string
	topicQueue map[string][]string
}

func NewAgent(logDir string) *Agent {
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chmod(logDir, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return &Agent{
		subs:       make(map[string][]chan string),
		quit:       make(chan struct{}),
		logDir:     logDir,
		topicQueue: make(map[string][]string),
	}
}

func (b *Agent) Publish(topic string, msg string) {
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

func (b *Agent) Subscribe(topic string) <-chan string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	ch := make(chan string)
	b.subs[topic] = append(b.subs[topic], ch)
	return ch
}

func (b *Agent) writeToFile(filePath string, msg string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(msg + "\n"); err != nil {
		return err
	}
	return nil
}

func (b *Agent) Close() {
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

func (b *Agent) saveTopicQueues() {
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
