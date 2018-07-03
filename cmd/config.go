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
	"text/template"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const configTemplate = `# MQTT configuration
mqtt:

  # MQTT server address
  #
  # The format should be scheme://host:port
  # Where "scheme" is one of "tcp", "ssl", or "ws", 
  # "host" is the ip-address (or hostname) and 
  # "port" is the port on which the broker is accepting connections.
  # 
  # Example:
  # server: tcp://foobar.com:1883
  server: {{viper "mqtt.server"}}

  # Username
  username: {{viper "mqtt.username"}}

  # Password
  password: {{viper "mqtt.password"}}

  # Client ID
  clientid: {{viper "mqtt.clientid"}}

  # Quality of Service
  qos: {{viper "mqtt.qos"}}

  # Topics
  topic: {{viper "mqtt.topic"}}

  # Debug enable
  debug: {{viper "mqtt.debug"}}

# DDR API
ddr:

  # DDR API url
  url: {{viper "ddr.url"}}
`

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the lora-ddr configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		funcMap := template.FuncMap{
			"viper": viper.GetString,
		}

		tmpl, err := template.New("config").Funcs(funcMap).Parse(configTemplate)
		if err != nil {
			log.Fatalf("[config] parsing: %s", err)
		}

		err = tmpl.Execute(os.Stdout, "config")
		if err != nil {
			log.Fatalf("[config] execution: %s", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(configCmd)

	// MQTT defaults
	viper.SetDefault("mqtt.server", "tcp://localhost:1883")
	viper.SetDefault("mqtt.username", "admin")
	viper.SetDefault("mqtt.password", "admin")
	viper.SetDefault("mqtt.clientid", "lora-mqtt-id")
	viper.SetDefault("mqtt.qos", 0)
	viper.SetDefault("mqtt.topic", "my_application/devices/+/up")
	viper.SetDefault("mqtt.debug", false)

	// DDR defaults
	viper.SetDefault("ddr.url", "http://localhost/lora/ddr/q")
}
