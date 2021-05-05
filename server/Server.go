package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/husobee/vestigo"

	"github.com/w8mr/gomoticasa/config"
	"github.com/w8mr/gomoticasa/controller"
)

const hist_size = 100

var context = Context{"Low_High", "Low", "Low", 0.0, 20.0, 0.0, time.Unix(0, 0), [hist_size]float64{}, 0}

var speeds = map[string](map[string]string){
	"Low_Low":       {"Low": "Low", "Medium": "Low", "High": "Low"},
	"Low_Medium":    {"Low": "Low", "Medium": "Medium", "High": "Medium"},
	"Low_High":      {"Low": "Low", "Medium": "Medium", "High": "High"},
	"Medium_Medium": {"Low": "Medium", "Medium": "Medium", "High": "Medium"},
	"Medium_High":   {"Low": "Medium", "Medium": "Medium", "High": "High"},
	"High_High":     {"Low": "High", "Medium": "High", "High": "High"},
}

type Context struct {
	mode        string
	speed       string
	fanspeed    string
	humidity    float64
	temperature float64
	switchHumidity    float64
	lastUpdated time.Time
	history     [hist_size]float64
	history_index int
}

//define a function for the default message handler
var bathroom_sensor mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Sensor message: %s\n", msg.Payload())

	var f interface{}
	err := json.Unmarshal(msg.Payload(), &f)
	if err != nil {
		panic(err)
	}
	m := f.(map[string]interface{})
	context.humidity = traverseJSONMap(m, "AM2301.Humidity").(float64)
	context.temperature = traverseJSONMap(m, "AM2301.Temperature").(float64)
	context.lastUpdated = time.Now()

	updateHistory(&context)

	oldFanspeed := context.fanspeed
	calcSpeed(&context)
	if oldFanspeed != context.fanspeed {
		context.switchHumidity = context.humidity
		log.Printf("Changed speed, because humidity changed, new speed: %v", context.fanspeed)
		setSpeed(client, &context)
	}
}

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Generic message: %s\n", msg.Payload())
}

var wtwMode mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("WTW message: %s\n", msg.Payload())
}


func updateHistory(context *Context) {
	index := context.history_index
	context.history[index] = context.humidity
	index = (index + 1) % hist_size
	context.history_index = index
}

func readAverageHistory(context *Context, history int, size int) float64{
	total := 0.0
	index := context.history_index
	for i:=0; i < size; i++ {
		total = total + context.history[(index - history - i + hist_size) % hist_size]
	}
	return total / float64(size)
}

func historyDiff(context *Context, diff int) float64 {
	return readAverageHistory(context, diff, 3) - readAverageHistory(context, 0, 3)
}

func calcSpeed(context *Context) {
	switch context.speed {
	case "Low":
		{
			if context.humidity > 70.0 {
				context.speed = "Medium"
			}
			if context.humidity > context.switchHumidity + 2.0 {
				context.speed = "Medium"
			}
		}
	case "Medium":
		{
			if context.humidity < 50.0 {
				context.speed = "Low"
			}
			if context.humidity > 80.0 {
				context.speed = "High"
			}
			if context.humidity < 69.0 {
				if historyDiff(context,10) < 0.2 {
					context.speed = "Low"
				}
			}
			if context.humidity > context.switchHumidity + 2.0 {
				context.speed = "High"
			}
		}
	case "High":
		{
			if context.humidity < 55.0 {
				context.speed = "Medium"
			}
			if context.humidity < 79.0 {
				if historyDiff(context,10) < 0.2 {
					context.speed = "Medium"
				}
			}
		}
	}
	context.fanspeed = speeds[context.mode][context.speed]
}

func setSpeed(client mqtt.Client, context *Context) {
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
}

func traverseJSONMap(m map[string]interface{}, path string) interface{} {
	parts := strings.SplitAfterN(path, ".", 2)
	key := strings.TrimSuffix(parts[0], ".")
	if len(parts) == 1 {
		return m[key]
	} else {
		return traverseJSONMap(m[key].(map[string]interface{}), parts[1])
	}
}

func Run(cfg *config.Config) error {
    mqttController := controller.NewMqttController(cfg)



	//subscribe to the topic /go-mqtt/sample and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription
	if token := mqttController.Client.Subscribe("tele/sonoff_bathroom/SENSOR", 0, bathroom_sensor); token.Wait() && token.Error() != nil {
		log.Println(fmt.Sprintf("Error subscribing sensor: %v", token.Error()))
		os.Exit(1)
	}

	log.Println("Subscribed")

	router := vestigo.NewRouter()
	//controller.SetupStatic(router)

	router.Get("/", modeHandler(mqttController.Client, &context))
	router.Post("/action", actionHandler(&context))

	setupTimer()


	http.Handle("/", router)
	log.Println(fmt.Sprintf("Server starting on port %d", cfg.Server.Port))
	address := fmt.Sprintf(":%d", cfg.Server.Port)
	err := http.ListenAndServe(
		address,
		nil)

	return err
}

func actionHandler(context *Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Print(r)
		_ = context
	}
}

func setupTimer() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for t := range ticker.C {
			if (context.lastUpdated.Unix() != 0) &&
				(time.Now().Sub(context.lastUpdated).Minutes() > 5.0) {
					_ = t
					log.Panicf("Exited because no event are recieved anymore")
			}
		}
	}()
}

func lastUpdated(context *Context) string {
	if context.lastUpdated.Unix() != 0 {
		return fmt.Sprintf("%.0f seconden geleden", time.Now().Sub(context.lastUpdated).Seconds())
	} else {
		return "Nooit"
	}
}

func modeHandler(client mqtt.Client, context *Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		modeParam := r.URL.Query().Get("mode")
		if speeds[modeParam] != nil {
			context.mode = modeParam
			log.Printf("Mode changed, new mode: %v", context.mode)

			oldFanspeed := context.fanspeed
			calcSpeed(context)
			if oldFanspeed != context.fanspeed {
				log.Printf("Changed speed, because mode changed, new speed: %v", context.fanspeed)
				setSpeed(client, context)
			}
		}
		handleView(context, w, r)
	}
}

func handleView(context *Context, w http.ResponseWriter, r *http.Request) {
	var selected = func(context *Context, value string) string {
		if context.mode == value {
			return "selected"
		} else {
			return ""
		}
	}

	fmt.Fprintf(w, "<html><head>"+
		"<meta name=viewport content=\"width=device-width, initial-scale=1, user-scalable=yes\">"+
		"<style>"+
		"body { font-family: arial; font-size:16px }"+
		".button { background: #008CBA; color: white; border-radius:12px; padding: 8px 16px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; margin: 4px 2px; -webkit-transition-duration: 0.4s; transition-duration: 0.4s; cursor: pointer; }"+
		".button:hover, .selected { background: white; border: 2px solid #008CBA }"+
		".selected { color: #008CBA }"+
		".button:hover, .selected:hover { color: black }"+
		"</style>"+
		"</head><body>"+
		"<p>Luchtvochtigeid: %.1f</p>"+
		"<p>Huidige snelheid: %v</p>"+
		"<p>Laatst gedupdate: %v</p>"+
		"<form action=\"/\" method=\"GET\">"+
		"<button name=\"mode\" value=\"Low_Low\" class=\"button %s\">Laag</button>"+
		"<button name=\"mode\" value=\"Low_High\" class=\"button %s\">Auto</button>"+
		"<button name=\"mode\" value=\"Medium_High\" class=\"button %s\">Middel</button>"+
		"<button name=\"mode\" value=\"High_High\" class=\"button %s\">Hoog</button>"+
		"</form>", context.humidity, context.fanspeed, lastUpdated(context), selected(context, "Low_Low"), selected(context, "Low_High"), selected(context, "Medium_High"), selected(context, "High_High"))
}
