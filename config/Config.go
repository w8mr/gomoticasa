package config

import (
	"github.com/namsral/flag"
)

type Config struct {
	Db     DbConfig
	Mqtt   MqttConfig
	Server ServerConfig
}


type DbConfig struct {
	Type     string
	Url     string
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
	flag.StringVar(&cfg.Db.Type, "dbtype", "stub", "Database type")
	flag.StringVar(&cfg.Db.Url, "dburl", "localhost:5432", "Database Url")

	flag.StringVar(&cfg.Mqtt.Url, "mqtturl", "tcp://192.168.1.181:1883", "Mqtt url")
	flag.StringVar(&cfg.Mqtt.User, "mqttuser", "openhab", "Mqtt user")
	flag.StringVar(&cfg.Mqtt.Password, "mqttpassword", "cafe123456", "Mqtt password")

	flag.IntVar(&cfg.Server.Port, "serverport", 8080, "Server port")


	flag.Parse()
}
