package stt

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

type STTMessage struct {
	Type     string  `json:"type"`
	Text     string  `json:"text,omitempty"`
	Language string  `json:"language,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Message  string  `json:"message,omitempty"`
}

func StreamToSTT(pcmChan <-chan []byte, url string, onText func(msg STTMessage)) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	go func() {
		defer func() {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("stop"))
		}()
		for chunk := range pcmChan {
			if len(chunk) == 0 {
				continue
			}
			_ = conn.WriteMessage(websocket.BinaryMessage, chunk)
		}
	}()

	go func() {
		for {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			time.Sleep(10 * time.Second)
		}
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var m STTMessage
		if err := json.Unmarshal(msgBytes, &m); err != nil {
			continue
		}
		if onText != nil {
			onText(m)
		}
	}
}
