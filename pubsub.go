package pubsub

import (
	"sync"
)

type Agent struct {
	Mu     sync.Mutex
	Subs   map[string][]chan string
	Quit   chan struct{}
	Closed bool
}

func NewAgent() *Agent {
	return &Agent{
		Subs: make(map[string][]chan string),
		Quit: make(chan struct{}),
	}
}

func (b *Agent) Publish(topic string, msg string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if b.Closed {
		return
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
}

// func main() {
// 	agent := NewAgent()
// 	sub := agent.Subscribe("a")

// 	go func() {
// 		for i := 0; i < 3; i++ {
// 			agent.Publish("a", fmt.Sprintf("Mensagem %d", i+1))
// 		}
// 	}()

// 	go func() {
// 		for i := 0; i < 3; i++ {
// 			fmt.Println(<-sub)
// 		}
// 	}()

// 	time.Sleep(time.Second)
// 	agent.Close()
// }
