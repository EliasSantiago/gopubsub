package pubsub

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestAgent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "pubsub_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	agent := NewAgent[any](tmpDir)
	ch := agent.Subscribe("noticias")
	go func() {
		for range ch {
		}
	}()
	agent.Publish("noticias", "Notícia 1")
	agent.Publish("noticias", "Notícia 2")
	agent.Close()
	logFilePath := tmpDir + "/noticias.txt"
	content, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		t.Fatal(err)
	}
	expectedContent := "Notícia 1\nNotícia 2\n"
	if string(content) != expectedContent {
		t.Errorf("Conteúdo do arquivo de log incorreto. Esperado: %s, Obtido: %s", expectedContent, string(content))
	}
}

func TestAgent_Publish(t *testing.T) {
	agent := NewAgent[any]("logs")
	ch := agent.Subscribe("noticias")
	outCh := make(chan string)
	go func() {
		for msg := range ch {
			fmt.Println(msg)
		}
		close(outCh)
	}()
	agent.Publish("noticias", "Notícia 1")
	agent.Publish("noticias", "Notícia 2")
	expectedMessages := []string{"Notícia 1", "Notícia 2"}
	receivedMessages := make([]string, 0)
	for msg := range outCh {
		receivedMessages = append(receivedMessages, msg)
	}
	if len(receivedMessages) != len(expectedMessages) {
		t.Errorf("Número incorreto de mensagens recebidas. Esperado: %d, Obtido: %d", len(expectedMessages), len(receivedMessages))
	}
	for i := 0; i < len(expectedMessages); i++ {
		if receivedMessages[i] != expectedMessages[i] {
			t.Errorf("Mensagem recebida incorreta. Índice: %d, Esperado: %s, Obtido: %s", i, expectedMessages[i], receivedMessages[i])
		}
	}
	agent.Close()
}

func TestAgent_Subscribe(t *testing.T) {
	agent := NewAgent[any]("logs")
	ch := agent.Subscribe("noticias")
	if ch == nil {
		t.Errorf("Falha ao criar canal de assinatura")
	}
	go func() {
		for range ch {
		}
	}()
	agent.Publish("noticias", "Notícia 1")
	agent.Publish("noticias", "Notícia 2")
	agent.Close()
}

func TestAgent_Close(t *testing.T) {
	agent := NewAgent[any]("logs")
	agent.Close()
	ch := agent.Subscribe("noticias")
	go func() {
		for range ch {
		}
	}()
	agent.Publish("noticias", "Notícia 1")
	agent.Publish("noticias", "Notícia 2")
	agent.Close()
}

func TestAgent_writeToFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "pubsub_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	agent := NewAgent[any](tmpDir)
	filePath := tmpDir + "/test.txt"
	msg := "Test Message"
	err = agent.writeToFile(filePath, msg)
	if err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != msg+"\n" {
		t.Errorf("Conteúdo do arquivo incorreto. Esperado: %s, Obtido: %s", msg+"\n", string(content))
	}
}
