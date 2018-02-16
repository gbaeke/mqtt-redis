package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var mqttClient MQTT.Client

// Location derived from MQTT topic
type Location struct {
	City     string
	Building string
	Room     string
}

// NewMQTTClient sets up the MQTT client
func NewMQTTClient(host string, port int, user string, password string, topic string) error {
	server := fmt.Sprintf("tcps://%s:%v", host, port)
	hostname, _ := os.Hostname()
	connOpts := MQTT.NewClientOptions()
	connOpts.AddBroker(server)
	connOpts.SetClientID(hostname)
	connOpts.SetCleanSession(true)
	connOpts.SetUsername(user)
	connOpts.SetPassword(password)
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	connOpts.SetTLSConfig(tlsConfig)
	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(topic, 0, onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	mqttClient = MQTT.NewClient(connOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	log.Println("MQTT client connection to ", server)
	return nil

}

// onMessageReceived is triggered by subscription on MQTTTopic (default #)
func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	log.Printf("Received message on MQTT topic: %s\n", message.Topic())

	//verify the MQTT topic; is it like airq/city/building/room?
	//if not, log the error and return; no data sent to Redis
	location, err := verifyTopic(message.Topic())
	if err != nil {
		log.Println(err)
		return
	}

	// write data to Redis
	// airq/city/building/room becomes airq:city:building:room
	channel := fmt.Sprintf("airq:%s:%s:%s", location.City, location.Building, location.Room)

	// note that message.Payload() is the original payload sent by the device to Mosquitto
	err = redisClient.Publish(channel, string(message.Payload()))
	if err != nil {
		// log the error but continue
		log.Println("Could not publish to Redis ", err)
	} else {
		log.Printf("Forwarded message to Redis on channel: %s\n", channel)
	}

}

// verifyTopic checks the MQTT topic conforms to airq/city/building/room
func verifyTopic(topic string) (*Location, error) {
	location := &Location{}
	items := strings.Split(topic, "/")
	if len(items) != 4 {
		return nil, errors.New("MQTT topic requires 4 sections: airq, city, building, room")
	}

	location.City = items[1]
	location.Building = items[2]
	location.Room = items[3]

	if items[0] != "airq" {
		return nil, errors.New("MQTT topic needs to start with airq")
	}

	if location.City == "" || location.Building == "" || location.Room == "" {
		return nil, errors.New("MQTT topic needs to to airq/city/building/room")
	}

	return location, nil
}
