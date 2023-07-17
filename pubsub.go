package pubsub

import (
	"log"
	"os"
	"sync"
)

type Agent struct {
	Mu         sync.Mutex
	Subs       map[string][]chan string
	Quit       chan struct{}
	Closed     bool
	LogDir     string
	TopicQueue map[string][]string
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
		Subs:       make(map[string][]chan string),
		Quit:       make(chan struct{}),
		LogDir:     logDir,
		TopicQueue: make(map[string][]string),
	}
}

func (b *Agent) Publish(topic string, msg string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if b.Closed {
		return
	}
	logFilePath := b.LogDir + "/" + topic + ".txt"
	b.TopicQueue[topic] = append(b.TopicQueue[topic], msg)
	if err := b.writeToFile(logFilePath, msg); err != nil {
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
	b.saveTopicQueues()
}

func (b *Agent) saveTopicQueues() {
	for topic, queue := range b.TopicQueue {
		filePath := b.LogDir + "/" + topic + ".txt"
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
