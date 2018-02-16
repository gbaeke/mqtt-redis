package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
)

// Specification for environment variables
type Specification struct {
	MQTTHost  string `envconfig:"MQTT_MOSQUITTO_SERVICE_HOST" required:"true"` // matches Kubernetes environment variable for service mqtt-mostquitto
	MQTTPort  int    `envconfig:"MQTT_MOSQUITTO_SERVICE_PORT" default:"8883"`
	MQTTUser  string `envconfig:"MQTT_USER" default:""`
	MQTTPass  string `envconfig:"MQTT_PASS" default:""`
	MQTTTopic string `envconfig:"MQTT_TOPIC" default:"airq/#"`
	RedisHost string `envconfig:"PUBSUB_REDIS_SERVICE_HOST" required:"true"` // matches Kubernetes environment variable for service db-influxdb
	RedisPort string `envconfig:"PUBSUB_REDIS_SERVICE_PORT" required:"true"`
}

var redisClient *RedisClient

// defined at package level; initialised in init(); used in main()
var s Specification

func init() {
	// get environment variables via Specification
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	log.Println("Setting up Redis client")
	redisAddress := fmt.Sprintf("%s:%s", s.RedisHost, s.RedisPort)
	redisOptions := &redis.Options{Addr: redisAddress}
	var err error
	redisClient, err = NewRedisClient(redis.NewClient(redisOptions))
	if err != nil {
		log.Println("Error during creation of Redis client ", err)
	}

	log.Println("Setting up MQTT client")
	err = NewMQTTClient(s.MQTTHost, s.MQTTPort, s.MQTTUser, s.MQTTPass, s.MQTTTopic)
	if err != nil {
		log.Println("Error during creation of  MQTT client ", err)
	}
	log.Println("Created MQTT client. Forwarding data...")

	// wait for SIGTERM
	<-c

}
