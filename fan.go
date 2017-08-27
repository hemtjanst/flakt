package main

import (
	"github.com/hemtjanst/hemtjanst/device"
	"github.com/hemtjanst/hemtjanst/messaging"
	"log"
	"strconv"
)

type Fan struct {
	announce      bool
	dev           *device.Device
	mq            messaging.PublishSubscriber
	currentActive bool
	currentSpeed  int
	targetSpeed   int
	currentSwing  int
	targetSwing   bool
	powerTopic    string
	speedTopic    string
	swingTopic    string
	speedFt       *device.Feature
	activeFt      *device.Feature
	swingFt       *device.Feature
}

func NewFan(mq messaging.PublishSubscriber, powerTopic, speedTopic, swingTopic string) *Fan {
	f := &Fan{
		announce:      false,
		mq:            mq,
		currentActive: false,
		currentSpeed:  1,
		targetSpeed:   33,
		currentSwing:  0,
		targetSwing:   false,
		powerTopic:    powerTopic,
		speedTopic:    speedTopic,
		swingTopic:    swingTopic,
	}
	return f
}

func (f *Fan) Start(d *device.Device) {
	f.dev = d
	d.Type = "fanV2"

	f.speedFt = &device.Feature{Step: 33}
	f.activeFt = &device.Feature{}
	f.swingFt = &device.Feature{}

	d.AddFeature("rotationSpeed", f.speedFt)
	d.AddFeature("active", f.activeFt)
	d.AddFeature("swingMode", f.swingFt)

	if f.announce {
		f.subscribeFeatures()
		d.PublishMeta()
	}
}

func (f *Fan) OnConnect() {
	go func() {
		f.mq.Subscribe("discover", 1, func(message messaging.Message) {
			go func() {
				f.announce = true
				if f.dev != nil {
					log.Print("Got discover, sending announce: " + *flgDeviceTopic)
					f.dev.PublishMeta()
				}
			}()
		})
		f.subscribeFeatures()
	}()
}

func (f *Fan) subscribeFeatures() {
	if f.swingFt != nil {
		f.swingFt.OnSet(func(msg messaging.Message) {
			log.Print("New swingMode received: " + string(msg.Payload()))
			f.SetSwing(string(msg.Payload()) == "1")
		})
	}
	if f.speedFt != nil {
		f.speedFt.OnSet(func(msg messaging.Message) {
			log.Print("New rotationSpeed received: " + string(msg.Payload()))
			if i, err := strconv.Atoi(string(msg.Payload())); err == nil {
				f.SetSpeed(i)
			}
		})
	}
	if f.activeFt != nil {
		f.activeFt.OnSet(func(msg messaging.Message) {
			log.Print("New state received: " + string(msg.Payload()))
			val, err := strconv.ParseBool(string(msg.Payload()))
			if err != nil {
				return
			}
			if val {
				f.StartFan()
			} else {
				f.StopFan()
			}
		})
	}
}

func (f *Fan) StartFan() {
	if !f.currentActive {
		f.mq.Publish(*flgPowerTopic, []byte("0"), 0, false)
		f.currentActive = true
		f.currentSpeed = 1
		f.currentSwing = 0
		f.SetSpeed(f.targetSpeed)
		f.SetSwing(f.targetSwing)
	}

	f.activeFt.Update("1")
}

func (f *Fan) StopFan() {
	if f.currentActive {
		f.mq.Publish(*flgPowerTopic, []byte("0"), 0, false)
		f.currentActive = false
	}

	f.activeFt.Update("0")
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
			f.mq.Publish(*flgSpeedTopic, []byte("0"), 0, false)
		}
	}
	f.currentSpeed = speedMode
	newSpeed = speedMode * 33
	if newSpeed == 99 {
		newSpeed = 100
	}
	f.speedFt.Update(strconv.Itoa(newSpeed))
}

func (f *Fan) SetSwing(swingMode bool) {
	f.targetSwing = swingMode
	if !f.currentActive {
		return
	}

	if (f.currentSwing == 1) != swingMode {
		f.mq.Publish(*flgSwingTopic, []byte("0"), 0, false)
		if swingMode {
			f.currentSwing = 1
		} else {
			f.currentSwing = 0
		}
	}
	f.swingFt.Update(strconv.Itoa(f.currentSwing))
}
