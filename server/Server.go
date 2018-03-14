package server

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/husobee/vestigo"
	"log"
	"net/http"
	"os"
	"strings"
	"w8mr.nl/go_my_home/config"
	"w8mr.nl/go_my_home/controller"
)

var mode = "Low_High"
var speed = "Low"
var fanspeed = "Low"
var humidity = 100.0;
var temperature = 20.0;


var speeds = map[string](map[string]string){
	"Low_Low":       {"Low": "Low", "Medium": "Low", "High": "Low"},
	"Low_Medium":    {"Low": "Low", "Medium": "Medium", "High": "Medium"},
	"Low_High":      {"Low": "Low", "Medium": "Medium", "High": "High"},
	"Medium_Medium": {"Low": "Medium", "Medium": "Medium", "High": "Medium"},
	"Medium_High":   {"Low": "Medium", "Medium": "Medium", "High": "High"},
	"High_High":     {"Low": "High", "Medium": "High", "High": "High"},
}

//define a function for the default message handler
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())

	var f interface{}
	err := json.Unmarshal(msg.Payload(), &f)
	if err != nil {
		panic(err)
	}
	m := f.(map[string]interface{})
	humidity = traverseJSONMap(m, "AM2301.Humidity").(float64)
	temperature = traverseJSONMap(m, "AM2301.Temperature").(float64)
	fmt.Printf("Humidity: %f\n", humidity)
	fmt.Printf("Temperature: %f\n", temperature)
	fmt.Printf("Speed before: %v\n", speed)

	switch speed {
	case "Low":
		{
			if humidity > 50.0 {
				speed = "Medium"
			}
		}
	case "Medium":
		{
			if humidity < 48.0 {
				speed = "Low"
			}
			if humidity > 67.0 {
				speed = "High"
			}
		}
	case "High":
		{
			if humidity < 65.0 {
				speed = "Medium"
			}
		}
	}

	fmt.Printf("Speed after: %v\n", speed)

	fanspeed = speeds[mode][speed]

	fmt.Printf("Fan speed: %v\n", fanspeed)

}

func traverseJSONMap(m map[string]interface{}, path string) interface{} {
	parts := strings.SplitAfterN(path, ".", 2)
	key := strings.TrimSuffix(parts[0], ".")
	if len(parts) == 1 {
		fmt.Printf("1 part: %s, %v\n", key, m[key])
		return m[key]
	} else {
		fmt.Printf("2 parts: %s, %v, %s\n", key, m[key], parts[1])
		return traverseJSONMap(m[key].(map[string]interface{}), parts[1])
	}
}

func Run(cfg *config.Config) error {

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := mqtt.NewClientOptions().AddBroker("tcp://192.168.1.180:1883")
	opts.SetClientID("go-my-home")
	opts.SetDefaultPublishHandler(f)
	opts.SetPassword("cafe123456")
	opts.SetUsername("openhab")

	//create and start a client using the above ClientOptions
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("Subsribe")
	//subscribe to the topic /go-mqtt/sample and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription
	if token := c.Subscribe("tele/sonoff_bathroom/SENSOR", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	log.Println("Message")
	text := fmt.Sprintf("this is msg #%d!", 1)
	token := c.Publish("tele/sonoff_bathroom/STATE", 0, false, text)
	token.Wait()

	log.Println("Done")
	//c.Disconnect(250)

	router := vestigo.NewRouter()
	controller.SetupStatic(router)

	http.Handle("/", router)
	log.Println(fmt.Sprintf("Server starting on port %d", cfg.Server.Port))
	address := fmt.Sprintf(":%d", cfg.Server.Port)
	err := http.ListenAndServe(
		address,
		nil)

	return err
}
