package controllerservice

import (
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type MqttClient struct {
	Client mqtt.Client
}

func NewMqttClient(
	broker string,
	port int,
	msgHandler mqtt.MessageHandler,
	onConnectHandler mqtt.OnConnectHandler,
	onConnectLostHandler mqtt.ConnectionLostHandler,
) *MqttClient {
	clientID := fmt.Sprintf("goSensorDataCollector-%s", uuid.New().String())
	mqttListenerOpts := mqtt.NewClientOptions()
	mqttListenerOpts.AddBroker(fmt.Sprintf("mqtt://%s:%d", broker, port))
	mqttListenerOpts.SetClientID(clientID)
	mqttListenerOpts.SetOrderMatters(false)
	mqttListenerOpts.SetDefaultPublishHandler(msgHandler)
	mqttListenerOpts.OnConnect = onConnectHandler
	mqttListenerOpts.OnConnectionLost = onConnectLostHandler
	mqttListenerOpts.SetKeepAlive(60 * time.Second)
	mqttListenerOpts.SetPingTimeout(10 * time.Second)

	newClient := createMqttClient(mqttListenerOpts)
	return &MqttClient{
		Client: newClient,
	}
}

func createMqttClient(opts *mqtt.ClientOptions) mqtt.Client {
	newClient := mqtt.NewClient(opts)
	if token := newClient.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("Error connecting to MQTT Broker", "error", token.Error())
		panic(token.Error())
	}
	return newClient
}
