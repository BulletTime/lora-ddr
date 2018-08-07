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
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/bullettime/lora-ddr/model/lora"
	"github.com/bullettime/lora-ddr/model/mqtt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

/*
 DDR Format:

 DDR communication is always sent over port 1

 Uplink: DDR|LAT|LON
 For example: DDR|50.863978|4.678908

 Downlink: DDR|SF
 For example: DDR|7
*/

const (
	DDRHEADER = "DDR"
	SEPARATOR = "|"

	UPLINKSUFFIX   = "up"
	DOWNLINKSUFFIX = "down"
)

type DDRResponse struct {
	Datarate string `json:"datarate"`
}

type LatLon struct {
	Latitude  float64
	Longitude float64
}

var client = mqtt.GetMQTT()

func getUplinkMessage(payload []byte) (*lora.UplinkMessage, error) {
	var uplink lora.UplinkMessage

	err := json.Unmarshal(payload, &uplink)
	if err != nil {
		return nil, err
	}

	return &uplink, nil
}

func isValidDDRRequest(payload []byte) bool {
	strPayload := string(payload)

	if !(strings.HasPrefix(strPayload, DDRHEADER) && strings.Contains(strPayload, SEPARATOR)) {
		return false
	}

	if strings.Count(strPayload, SEPARATOR) != 2 {
		return false
	}

	return true
}

func getLatLon(payload []byte) (*LatLon, error) {
	request := strings.Split(string(payload), SEPARATOR)

	if len(request) != 3 {
		return nil, errors.New("invalid ddr request")
	}

	lat, err := strconv.ParseFloat(request[1], 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid latitude in ddr request")
	}

	lon, err := strconv.ParseFloat(request[2], 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid longitude in ddr request")
	}

	return &LatLon{Latitude: lat, Longitude: lon}, nil
}

func queryDDR(coordinates *LatLon) (*DDRResponse, error) {
	ddrUrl, err := url.Parse(viper.GetString("ddr.url"))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse ddr url")
	}

	query := ddrUrl.Query()
	query.Add("lat", strconv.FormatFloat(coordinates.Latitude, 'f', 6, 64))
	query.Add("lon", strconv.FormatFloat(coordinates.Longitude, 'f', 6, 64))
	ddrUrl.RawQuery = query.Encode()

	response, err := http.Get(ddrUrl.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not call ddr api")
	}
	defer response.Body.Close()

	var ddrResponse DDRResponse

	if err := json.NewDecoder(response.Body).Decode(&ddrResponse); err != nil {
		return nil, errors.Wrap(err, "could not decode ddr response")
	}

	return &ddrResponse, nil
}

func generateDownlinkMessage(response *DDRResponse) (*lora.DownlinkMessage, error) {
	var buffer bytes.Buffer
	buffer.WriteString(DDRHEADER)
	buffer.WriteString(SEPARATOR)
	// TODO Convert to int to check bounds if untrusted before adding it to the buffer
	buffer.WriteString(response.Datarate[2 : len(response.Datarate)-5])

	downlink := lora.DownlinkMessage{
		Port:      1,
		Confirmed: true,
		Payload:   buffer.Bytes(),
	}

	return &downlink, nil
}

func generateDownlinkTopic(uplinkTopic string) (string, error) {
	topic := uplinkTopic
	if !strings.HasSuffix(topic, UPLINKSUFFIX) {
		return "", errors.New("invalid uplink topic")
	}

	topic = strings.TrimSuffix(topic, UPLINKSUFFIX)
	topic = strings.Join([]string{topic, DOWNLINKSUFFIX}, "")

	return topic, nil
}

func MessageHandler(_ paho.Client, message paho.Message) {
	uplink, err := getUplinkMessage(message.Payload())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"topic":   message.Topic(),
			"payload": message.Payload(),
		}).Warn("can't unmarshall payload to uplink message")
		return
	}

	if !isValidDDRRequest(uplink.Payload) {
		log.WithFields(log.Fields{
			"topic":  message.Topic(),
			"uplink": uplink,
		}).Debug("received message other than ddr request")
		return
	}

	coordinates, err := getLatLon(uplink.Payload)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"topic":   message.Topic(),
			"payload": message.Payload(),
		}).Warn("can't retrieve coordinates from ddr request")
		return
	}

	response, err := queryDDR(coordinates)
	if err != nil {
		log.WithError(err).WithField("coordinates", coordinates).Warn("ddr request failed")
		return
	}

	downlink, err := generateDownlinkMessage(response)
	if err != nil {
		log.WithError(err).Warn("could not generate downlink message")
		return
	}

	topic, err := generateDownlinkTopic(message.Topic())
	if err != nil {
		log.WithError(err).WithField("uplink topic", message.Topic()).
			Warn("could not generate downlink topic")
	}

	log.WithFields(log.Fields{
		"topic":    topic,
		"downlink": downlink,
	}).Debug("set ddr")

	client.Publish(topic, downlink)
}
