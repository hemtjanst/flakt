package main

import (
	"flag"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hemtjanst/hemtjanst/device"
	"github.com/hemtjanst/hemtjanst/messaging"
	"github.com/hemtjanst/hemtjanst/messaging/flagmqtt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	flgPowerTopic   = flag.String("topic.power", "remote/fan/KEY_POWER", "Topic to send to when device is triggered")
	flgSpeedTopic   = flag.String("topic.speed", "remote/fan/KEY_SPEED", "Topic for changing speed")
	flgSwingTopic   = flag.String("topic.swing", "remote/fan/KEY_OSC", "Topic for changing swing mode")
	flgDeviceName   = flag.String("device.name", "Fan", "Device Name")
	flgDeviceModel  = flag.String("device.model", "", "Device Model")
	flgDeviceManu   = flag.String("device.manufacturer", "", "Device Manufacturer")
	flgDeviceSerial = flag.String("device.serial", "", "Device Serial number")
	flgDeviceTopic  = flag.String("topic.device", "switch/livingroom/fan", "Device topic")
)

func main() {

	var fan *Fan

	flag.Parse()

	id := flagmqtt.NewUniqueIdentifier()

	mqClient, err := flagmqtt.NewPersistentMqtt(flagmqtt.ClientConfig{
		WillTopic:   "leave",
		WillPayload: id,
		WillRetain:  false,
		WillQoS:     1,
		ClientID:    id,
		OnConnectHandler: func(client mqtt.Client) {
			if fan != nil {
				fan.OnConnect()
			}
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	mq := messaging.NewMQTTMessenger(mqClient)
	fan = NewFan(mq, *flgPowerTopic, *flgSpeedTopic, *flgSwingTopic)
	dev := device.NewDevice(*flgDeviceTopic, mq)
	dev.Name = *flgDeviceName
	dev.Manufacturer = *flgDeviceManu
	dev.Model = *flgDeviceModel
	dev.SerialNumber = *flgDeviceSerial
	dev.LastWillID = id
	fan.Start(dev)

	log.Print("Connecting to MQTT")
	token := mqClient.Connect()
	token.Wait()
	if token.Error() != nil {
		log.Fatal(err)
	}
	log.Print("Connected")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

}
