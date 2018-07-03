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

package mqtt

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/apex/log"
	"github.com/bullettime/lora-ddr/model/lora"
	"github.com/bullettime/lora-ddr/util"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type MQTT interface {
	Connect() error
	Subscribe(string, paho.MessageHandler) error
	Publish(string, *lora.DownlinkMessage) error
	Unsubscribe(...string) error
	Disconnect() error
}

type mqtt struct {
	client        paho.Client
	options       *paho.ClientOptions
	qos           byte
	subscriptions []string

	sync.Mutex
}

type debugLogger struct{}

var client = mqtt{}

func GetMQTT() MQTT {
	return &client
}

func (m *mqtt) Connect() error {
	m.Lock()
	defer m.Unlock()

	m.options = paho.NewClientOptions()

	m.options.AddBroker(viper.GetString("mqtt.server"))
	m.options.SetUsername(viper.GetString("mqtt.username"))
	m.options.SetPassword(viper.GetString("mqtt.password"))
	m.options.SetClientID(viper.GetString("mqtt.clientid"))

	m.qos = byte(viper.GetInt("mqtt.qos"))

	if viper.GetBool("mqtt.debug") {
		paho.DEBUG = debugLogger{}
	}

	m.client = paho.NewClient(m.options)

	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return errors.Wrap(token.Error(), "[MQTT] error connecting")
	}

	m.subscriptions = make([]string, 0)

	log.Info("[MQTT] connected")

	return nil
}

func (m *mqtt) Subscribe(topic string, callback paho.MessageHandler) error {
	if m.client != nil && m.client.IsConnected() {
		if token := m.client.Subscribe(topic, m.qos, callback); token.Wait() && token.Error() != nil {
			return errors.Wrapf(token.Error(), "[MQTT] error subscribing to: %s", topic)
		}

		m.subscriptions = append(m.subscriptions, topic)
		log.Infof("[MQTT] subscribed to topic: %s", topic)

		return nil
	}
	return errors.New("[MQTT] trying to subscribe while not connected")
}

func (m *mqtt) Publish(topic string, msg *lora.DownlinkMessage) error {
	js, err := json.Marshal(msg)
	if err != nil {
		return errors.New(fmt.Sprintf("[MQTT] error publishing to topic '%s' with message: %v", topic, msg))
	}

	m.client.Publish(topic, m.qos, false, js)

	return nil
}

func (m *mqtt) Unsubscribe(topics ...string) error {
	if m.client != nil && m.client.IsConnected() {
		if token := m.client.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
			return errors.Wrapf(token.Error(), "[MQTT] error unsubscribing from: %s", topics)
		}

		for _, topic := range topics {
			util.Remove(m.subscriptions, topic)
			log.Infof("[MQTT] unsubscribing from topic: %s", topic)
		}

		return nil
	}
	return errors.New("[MQTT] trying to unsubscribe while not connected")
}

func (m *mqtt) Disconnect() error {
	m.Lock()
	defer m.Unlock()

	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(250)
	}

	log.Info("[MQTT] disconnected")

	return nil
}

func (l debugLogger) Println(v ...interface{}) {
	log.Debugf("[MQTT Debug] %s", fmt.Sprintln(v...))
}

func (l debugLogger) Printf(format string, v ...interface{}) {
	log.Debugf("[MQTT Debug] %s", fmt.Sprintf(format, v...))
}
