package main

import (
	"github.com/ninjasphere/go-ninja/api"
	nconfig "github.com/ninjasphere/go-ninja/config"
)

func Mqtt() {

	conn, err := ninja.Connect("sphere-director")
	if err != nil {
		log.Errorf("Failed to connect to sphere: %s", err)
		panic("MQTT Fail!")
	}

	topic := "$node/" + nconfig.Serial() + "/module/:task"

	log.Infof("Subscribing to %s", topic)

	if _, err := conn.Subscribe("$node/"+nconfig.Serial()+"/module/start", func(name *string, data map[string]string) bool {
		log.Infof("Received request to start process %s", *name)

		p, ok := daemon.children[*name]
		if !ok {
			log.Infof("%s does not exist.", *name)
			return true
		}
		cp, _, _ := p.find()
		if cp != nil {
			log.Infof("%s already running.", *name)
			return true
		}
		ch := RunProcess(p)
		log.Debugf("%s", <-ch)

		return true
	}); err != nil {
		log.Fatalf("Failed to subscribe to module topic: %s", err)
	}

	if _, err := conn.Subscribe("$node/"+nconfig.Serial()+"/module/stop", func(name *string, data map[string]string) bool {
		log.Infof("Received request to stop process %s", *name)

		p, ok := daemon.children[*name]
		if !ok {
			log.Infof("%s does not exist.", *name)
			return true
		}

		p.find()
		p.stop()

		return true
	}); err != nil {
		log.Fatalf("Failed to subscribe to module topic: %s", err)
	}

	if _, err := conn.Subscribe("$node/"+nconfig.Serial()+"/module/restart", func(name *string, data map[string]string) bool {
		log.Infof("Received request to restart process %s", *name)

		p, ok := daemon.children[*name]
		if !ok {
			log.Infof("%s does not exist.", *name)
			return true
		}

		p.find()
		ch, _ := p.restart()
		log.Debugf("%s", <-ch)

		return true
	}); err != nil {
		log.Fatalf("Failed to subscribe to module topic: %s", err)
	}

}
