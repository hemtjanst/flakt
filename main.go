package main

import (
	"context"
	"flag"
	"github.com/hemtjanst/bibliotek/transport/mqtt"
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

	mqCfg := mqtt.MustFlags(flag.String, flag.Bool)
	var fan *Fan

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
	}()

	mq, err := mqtt.New(ctx, mqCfg())
	if err != nil {
		log.Fatal(err)
	}

	fan = NewFan(Config{
		PowerTopic:         *flgPowerTopic,
		SpeedTopic:         *flgSpeedTopic,
		SwingTopic:         *flgSwingTopic,
		DeviceTopic:        *flgDeviceTopic,
		DeviceName:         *flgDeviceName,
		DeviceModel:        *flgDeviceModel,
		DeviceManufacturer: *flgDeviceManu,
		DeviceSerial:       *flgDeviceSerial,
	})

	err = fan.Start(mq)
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
