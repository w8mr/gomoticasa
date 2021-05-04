package config

import (
	"github.com/namsral/flag"
)

type Config struct {
	Mqtt   MqttConfig
	Server ServerConfig
}

type MqttConfig struct {
	Url      string
	User     string
	Password string
}

type ServerConfig struct {
	Port int
}


func New() *Config {
	cfg := &Config{}
	return cfg
}

func (cfg *Config) FromFlags() {
	flag.StringVar(&cfg.Mqtt.Url, "MQTT_URL", "tcp://localhost:1883", "Mqtt url")
	flag.StringVar(&cfg.Mqtt.User, "MQTT_USER", "", "Mqtt user")
	flag.StringVar(&cfg.Mqtt.Password, "MQTT_PASSWORD", "", "Mqtt password")

	flag.IntVar(&cfg.Server.Port, "SERVER_PORT", 8080, "Server port")

	flag.Parse()
}
