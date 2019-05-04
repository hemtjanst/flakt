package main

import (
	"github.com/hemtjanst/bibliotek/client"
	"github.com/hemtjanst/bibliotek/device"
	"github.com/hemtjanst/bibliotek/feature"
	"github.com/hemtjanst/bibliotek/transport/mqtt"
	"log"
	"strconv"
)

type Config struct {
	PowerTopic         string
	SpeedTopic         string
	SwingTopic         string
	DeviceTopic        string
	DeviceName         string
	DeviceModel        string
	DeviceManufacturer string
	DeviceSerial       string
}

type Fan struct {
	dev           client.Device
	mq            mqtt.MQTT
	cfg           Config
	currentActive bool
	currentSpeed  int
	targetSpeed   int
	currentSwing  int
	targetSwing   bool
	speedFt       client.Feature
	activeFt      client.Feature
	swingFt       client.Feature
}

func NewFan(cfg Config) *Fan {
	f := &Fan{
		currentActive: false,
		currentSpeed:  1,
		targetSpeed:   33,
		currentSwing:  0,
		targetSwing:   false,
		cfg:           cfg,
	}
	return f
}

func (f *Fan) Start(tr mqtt.MQTT) error {
	dev, err := client.NewDevice(
		&device.Info{
			Topic:        f.cfg.DeviceTopic,
			Name:         f.cfg.DeviceName,
			Manufacturer: f.cfg.DeviceManufacturer,
			Model:        f.cfg.DeviceModel,
			SerialNumber: f.cfg.DeviceSerial,
			Type:         "fanV2",
			Features: map[string]*feature.Info{
				"rotationSpeed": {},
				"active":        {},
				"swingMode":     {},
			},
		},
		tr,
	)
	if err != nil {
		return err
	}
	f.mq = tr
	f.dev = dev
	f.speedFt = dev.Feature("rotationSpeed")
	_ = f.speedFt.OnSetFunc(func(msg string) {
		log.Print("New rotationSpeed received: " + msg)
		if i, err := strconv.Atoi(msg); err == nil {
			f.SetSpeed(i)
		}
	})

	f.activeFt = dev.Feature("active")
	_ = f.activeFt.OnSetFunc(func(msg string) {
		log.Print("New state received: " + msg)
		val, err := strconv.ParseBool(msg)
		if err != nil {
			return
		}
		if val {
			f.StartFan()
		} else {
			f.StopFan()
		}
	})

	f.swingFt = dev.Feature("swingMode")
	_ = f.swingFt.OnSetFunc(func(msg string) {
		log.Print("New swingMode received: " + msg)
		f.SetSwing(msg == "1")
	})

	return nil
}

func (f *Fan) StartFan() {
	if !f.currentActive {
		f.mq.Publish(f.cfg.PowerTopic, []byte("0"), false)
		f.currentActive = true
		f.currentSpeed = 1
		f.currentSwing = 0
		f.SetSpeed(f.targetSpeed)
		f.SetSwing(f.targetSwing)
	}

	_ = f.activeFt.Update("1")
}

func (f *Fan) StopFan() {
	if f.currentActive {
		f.mq.Publish(f.cfg.PowerTopic, []byte("0"), false)
		f.currentActive = false
	}

	_ = f.activeFt.Update("0")
}

func (f *Fan) SetSpeed(newSpeed int) {
	if newSpeed == 0 {
		f.StopFan()
		return
	}
	f.targetSpeed = newSpeed
	if !f.currentActive {
		return
	}

	speedMode := int(newSpeed / 33)
	n := speedMode - f.currentSpeed
	if n < 0 {
		n += 3
	}
	if n > 0 {
		for i := 0; i < n; i++ {
			f.mq.Publish(f.cfg.SpeedTopic, []byte("0"), false)
		}
	}
	f.currentSpeed = speedMode
	newSpeed = speedMode * 33
	if newSpeed == 99 {
		newSpeed = 100
	}
	_ = f.speedFt.Update(strconv.Itoa(newSpeed))
}

func (f *Fan) SetSwing(swingMode bool) {
	f.targetSwing = swingMode
	if !f.currentActive {
		return
	}

	if (f.currentSwing == 1) != swingMode {
		f.mq.Publish(f.cfg.SwingTopic, []byte("0"), false)
		if swingMode {
			f.currentSwing = 1
		} else {
			f.currentSwing = 0
		}
	}
	_ = f.swingFt.Update(strconv.Itoa(f.currentSwing))
}
