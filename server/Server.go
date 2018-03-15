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

var context = Context{"Low_High", "Low", "Low", 0.0, 20.0}

var speeds = map[string](map[string]string){
	"Low_Low":       {"Low": "Low", "Medium": "Low", "High": "Low"},
	"Low_Medium":    {"Low": "Low", "Medium": "Medium", "High": "Medium"},
	"Low_High":      {"Low": "Low", "Medium": "Medium", "High": "High"},
	"Medium_Medium": {"Low": "Medium", "Medium": "Medium", "High": "Medium"},
	"Medium_High":   {"Low": "Medium", "Medium": "Medium", "High": "High"},
	"High_High":     {"Low": "High", "Medium": "High", "High": "High"},
}

type Context struct {
	mode string
	speed string
	fanspeed string
	humidity float64
	temperature float64
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
	context.humidity = traverseJSONMap(m, "AM2301.Humidity").(float64)
	context.temperature = traverseJSONMap(m, "AM2301.Temperature").(float64)
	fmt.Printf("Humidity: %f\n", context.humidity)
	fmt.Printf("Temperature: %f\n", context.temperature)
	fmt.Printf("Speed before: %v\n", context.speed)

	calcSpeed(&context)
	setSpeed(client, &context)

	fmt.Printf("Speed after: %v\n", context.speed)

	}

func calcSpeed(context *Context) {
	switch context.speed {
	case "Low":
		{
			if context.humidity > 50.0 {
				context.speed = "Medium"
			}
		}
	case "Medium":
		{
			if context.humidity < 48.0 {
				context.speed = "Low"
			}
			if context.humidity > 67.0 {
				context.speed = "High"
			}
		}
	case "High":
		{
			if context.humidity < 65.0 {
				context.speed = "Medium"
			}
		}
	}

	context.fanspeed = speeds[context.mode][context.speed]
	fmt.Printf("Fan speed: %v\n", context.fanspeed)
}

func setSpeed(client mqtt.Client, context *Context) {
	log.Println("Message")

	var speed1 = "OFF"
	if context.fanspeed == "Medium" {
		speed1 = "ON"
	}
	token1 := client.Publish("cmnd/sonoff_wtw/power1", 0, false, speed1)

	var speed2 = "OFF"
	if context.fanspeed == "High" {
		speed2 = "ON"
	}
	token2 := client.Publish("cmnd/sonoff_wtw/power2", 0, false, speed2)

	token1.Wait()
	token2.Wait()

	log.Println("Done")

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

	//c.Disconnect(250)

	router := vestigo.NewRouter()
	controller.SetupStatic(router)

	router.Get("/mode", modeHandler(c, &context))


	http.Handle("/", router)
	log.Println(fmt.Sprintf("Server starting on port %d", cfg.Server.Port))
	address := fmt.Sprintf(":%d", cfg.Server.Port)
	err := http.ListenAndServe(
		address,
		nil)

	return err
}

func modeHandler(client mqtt.Client, context *Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		modeParam := r.URL.Query().Get("mode")
		fmt.Fprintf(w, "OK, mode=%v", modeParam)
		if speeds[modeParam] != nil {
			context.mode = modeParam
			calcSpeed(context)
			setSpeed(client, context)
		}
	}
}