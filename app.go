package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shopspring/decimal"
	"github.com/tkanos/gonfig"
	"log"
	"net/http"
	"os"
	"strings"
)

var gauges = make(map[string]prometheus.Gauge)
var minimalRegistry = prometheus.NewRegistry()

type Configuration struct {
	MqttUrl   string
	MqttTopic string
}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		panic(err)
	}

	mqttURL := configuration.MqttUrl
	topic := configuration.MqttTopic

	messageHandler := func(client mqtt.Client, message mqtt.Message) {
		payload := strings.TrimSpace(string(message.Payload()))
		split := strings.Split(payload, ":")
		if len(split) != 2 {
			println("Rejected malformed message: '" + payload + "'")
			return
		}

		name := normaliseMetricName(split[0])

		valueToParse := split[1]
		value, err := parseMaybeNumber(valueToParse)
		if err != nil {
			println("Rejected unparsable value: '" + valueToParse + "'")
		}

		gauge := getOrRegisterGauge(name)

		f, _ := value.Float64()
		gauge.Set(f)
	}

	mqttClient := setupMqttClient(mqttURL, "mqtt-scraper", topic, messageHandler)
	defer mqttClient.Disconnect(250)

	handler := promhttp.HandlerFor(minimalRegistry, promhttp.HandlerOpts{})
	http.Handle("/", handler)
	http.ListenAndServe(":8080", nil)
}

func normaliseMetricName(input string) string {
	// TODO implement correctly
	name := strings.ReplaceAll(input, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, ".", "")
	return name
}

func parseMaybeNumber(valueToParse string) (*decimal.Decimal, error) {
	value, err := decimal.NewFromString(valueToParse)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func getOrRegisterGauge(name string) prometheus.Gauge {
	gauge, found := gauges[name]
	if !found {
		gauge = promauto.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: "",
		})
		println("Registering new gauge: " + name)
		minimalRegistry.MustRegister(gauge)
		gauges[name] = gauge
	}
	return gauge
}

func setupMqttClient(mqttURL string, clientId string, topic string, handler mqtt.MessageHandler) mqtt.Client {
	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)

	var subscribeToTopic mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Print("Connected to " + mqttURL)
		log.Print("Subscribing to " + topic)
		client.Subscribe(topic, 0, handler)
	}
	var logConnectionLost mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		log.Print("Connection lost")
	}
	var logReconnecting mqtt.ReconnectHandler = func(client mqtt.Client, opts *mqtt.ClientOptions) {
		log.Print("Reconnecting")
	}

	opts := mqtt.NewClientOptions().AddBroker(mqttURL)
	opts.SetClientID(clientId)
	opts.SetOnConnectHandler(subscribeToTopic)
	opts.SetConnectionLostHandler(logConnectionLost)
	opts.SetReconnectingHandler(logReconnecting)

	println("Connecting to: ", mqttURL)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}
