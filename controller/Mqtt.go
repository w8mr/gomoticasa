package controller

import (
	"fmt"
	"log"
	"github.com/eclipse/paho.mqtt.golang"

	"github.com/w8mr/gomoticasa/config"
)

type MqttController struct {
	Client  mqtt.Client
	Token   mqtt.Token
}

func NewMqttController(cfg *config.Config) *MqttController {

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := mqtt.NewClientOptions().AddBroker(cfg.Mqtt.Url)
	opts.SetClientID("go-my-home")
	opts.SetDefaultPublishHandler(defaultHandler)
	opts.SetPassword(cfg.Mqtt.Password)
	opts.SetUsername(cfg.Mqtt.User)

	//create and start a client using the above ClientOptions
	client := mqtt.NewClient(opts)

	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}

	log.Println(fmt.Sprintf("Subscribe (url: %v, user: %v, password: %v)", cfg.Mqtt.Url, cfg.Mqtt.User, cfg.Mqtt.Password))

    mqttController := &MqttController{}
    mqttController.Client = client
    mqttController.Token = token

    return mqttController
}

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Generic message: %s\n", msg.Payload())
}
