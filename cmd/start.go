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

package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/apex/log"
	"github.com/bullettime/lora-ddr/model/ddr"
	"github.com/bullettime/lora-ddr/model/mqtt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ddr monitor",
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		client := mqtt.GetMQTT()

		if err := client.Connect(); err != nil {
			log.WithError(err).Fatal("cannot connect to mqtt")
		}

		if err := client.Subscribe(viper.GetString("mqtt.topic"), ddr.MessageHandler); err != nil {
			log.WithError(err).Fatal("cannot subscribe to topic")
		}

		waitForSignal()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}

func checkConfig() {
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("no config file found (run 'lora-ddr config > ~/.lora-ddr.yaml')")
	}
}

func waitForSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	log.WithField("signal", s).Warn("exiting")
}
