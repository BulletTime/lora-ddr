// The MIT License (MIT)
//
// Copyright © 2018 Sven Agneessens <sven.agneessens@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ddr

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bullettime/lora-ddr/model/lora"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
)

type mockMQTT struct{}

func (m *mockMQTT) Connect() error {
	return nil
}

func (m *mockMQTT) Subscribe(topic string, callback paho.MessageHandler) error {
	return nil
}

func (m *mockMQTT) Publish(topic string, msg *lora.DownlinkMessage) error {
	fmt.Printf("Publish:\n\ttopic: %s\n\tmessage:\n", topic)
	fmt.Printf("\t\tport: %v\n\t\tconfirmed: %v\n\t\tpayload: %s\n", msg.Port, msg.Confirmed, string(msg.Payload))
	return nil
}

func (m *mockMQTT) Unsubscribe(topics ...string) error {
	return nil
}

func (m *mockMQTT) Disconnect() error {
	return nil
}

type mockMessage struct {
	duplicate bool
	qos       byte
	retained  bool
	topic     string
	messageID uint16
	payload   []byte
}

func (m *mockMessage) Duplicate() bool {
	return m.duplicate
}

func (m *mockMessage) Qos() byte {
	return m.qos
}

func (m *mockMessage) Retained() bool {
	return m.retained
}

func (m *mockMessage) Topic() string {
	return m.topic
}

func (m *mockMessage) MessageID() uint16 {
	return m.messageID
}

func (m *mockMessage) Payload() []byte {
	return m.payload
}

func TestMessageHandler(t *testing.T) {
	oldClient := client
	defer func() { client = oldClient }()

	viper.SetDefault("ddr.url", "http://192.168.1.7:8080/lora/ddr/q")

	client = &mockMQTT{}

	uplink := lora.UplinkMessage{
		Port:    1,
		Payload: []byte("DDR|50.858955|4.671778"),
	}

	payload, _ := json.Marshal(uplink)

	message := &mockMessage{
		duplicate: false,
		qos:       0,
		retained:  false,
		topic:     "appid/devices/device1/up",
		messageID: 0,
		payload:   payload,
	}

	MessageHandler(nil, message)
}
