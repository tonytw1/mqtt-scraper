package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var gauges = make(map[string]prometheus.Gauge)
var minimalRegistry = prometheus.NewRegistry()

func main() {
	mqttURL := "tcp://mosquitto:1883"
	topic := "abbottroad"

	messageHandler := func(client mqtt.Client, message mqtt.Message) {
		payload := string(message.Payload())
		split := strings.Split(payload, ":")
		if len(split) != 2 {
			return
		}

		name := normaliseMetricName(split[0])

		value, err := strconv.ParseFloat(split[1], 32)
		if err != nil {
			return
		}

		gauge := getOrRegisterGauge(name)

		println("Setting: " + name)
		gauge.Set(value)
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

	var logConnection mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Print("Connected")
	}
	var logConnectionLost mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		log.Print("Connection lost")
	}
	var logReconnecting mqtt.ReconnectHandler = func(client mqtt.Client, opts *mqtt.ClientOptions) {
		log.Print("Reconnecting")
	}
	var subscribeToTopic mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Print("Subscribing")
		client.Subscribe(topic, 0, handler)
	}

	opts := mqtt.NewClientOptions().AddBroker(mqttURL)
	opts.SetOnConnectHandler(logConnection)
	opts.SetConnectionLostHandler(logConnectionLost)
	opts.SetReconnectingHandler(logReconnecting)
	opts.SetClientID(clientId)
	opts.SetOnConnectHandler(subscribeToTopic)

	println("Connecting to: ", mqttURL)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}
