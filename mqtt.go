package main

import (
	"github.com/ninjasphere/go-ninja/api"
	nconfig "github.com/ninjasphere/go-ninja/config"
)

func Mqtt() {
	log.Infof("WOOOT")

	conn, err := ninja.Connect("sphere-go-homecloud")
	if err != nil {
		log.Errorf("Failed to connect to sphere: %s", err)
		panic("MQTT Fail!")
	}

	topic := "$node/" + nconfig.Serial() + "/module/:task"

	log.Infof("Subscribing to %s", topic)

	if conn.Subscribe("$node/"+nconfig.Serial()+"/module/start", func(name *string, data map[string]string) bool {
		log.Infof("Received request to start process %s", *name)

		p := daemon.children.get(*name)
		if p == nil {
			log.Infof("%s does not exist.", *name)
			return true
		}
		cp, _, _ := p.find()
		if cp != nil {
			log.Infof("%s already running.", *name)
			return true
		}
		ch := RunProcess(*name, p)
		log.Debugf("%s", <-ch)

		return true
	}) != nil {
		log.Fatalf("Failed to subscribe to module topic")
	}

	if conn.Subscribe("$node/"+nconfig.Serial()+"/module/stop", func(name *string, data map[string]string) bool {
		log.Infof("Received request to stop process %s", *name)

		p := daemon.children.get(*name)
		if p == nil {
			log.Infof("%s does not exist.", *name)
			return true
		}

		p.find()
		p.stop()

		return true
	}) != nil {
		log.Fatalf("Failed to subscribe to module topic")
	}

	if conn.Subscribe("$node/"+nconfig.Serial()+"/module/restart", func(name *string, data map[string]string) bool {
		log.Infof("Received request to restart process %s", *name)

		p := daemon.children.get(*name)
		if p == nil {
			log.Infof("%s does not exist.", *name)
			return true
		}

		p.find()
		ch, _ := p.restart()
		log.Debugf("%s", <-ch)

		return true
	}) != nil {
		log.Fatalf("Failed to subscribe to module topic")
	}

}
