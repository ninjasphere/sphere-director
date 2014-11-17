// goforever - processes management
// Copyright (c) 2013 Garrett Woodworth (https://github.com/gwoo).

// sphere-director - Ninja processes management
// Copyright (c) 2014 Ninja Blocks Inc. (https://github.com/ninjablocks).

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	nconfig "github.com/ninjasphere/go-ninja/config"
)

type Config struct {
	Port      int
	Username  string
	Password  string
	Daemonize bool
	Pidfile   Pidfile
	Logfile   string
	Errfile   string
	Paths     []string
	Processes map[string]*Process
}

type packageJson struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Main        string `json:"main"`
	Author      string `json:"author"`
	License     string `json:"license"`
	MaxMemory   int    `json:"maxMemory"`
	Respawn     int    `json:"respawn"`
	Delay       string `json:"delay"`
	Ping        string `json:"ping"`
}

func (c Config) Keys() []string {
	keys := []string{}
	for k := range c.Processes {
		keys = append(keys, k)
	}

	spew.Dump("Keys", keys)
	return keys
}

func (c Config) Get(key string) *Process {
	return c.Processes[key]
}

func (c Config) FindProcesses() {

	for _, path := range c.Paths {
		path = os.ExpandEnv(path)

		log.Debugf("Finding processes in path %s", path)

		files, err := ioutil.ReadDir(path)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					//spew.Dump(file)

					infoFile, err := ioutil.ReadFile(filepath.Join(path, file.Name(), "package.json"))
					if err == nil {

						var info packageJson
						err = json.Unmarshal(infoFile, &info)

						if err != nil {
							log.Warningf("Could not read package.info for module %s : %s", file.Name(), err)
						} else {
							//spew.Dump(info)
							process := &Process{
								Name:        file.Name(),
								DisplayName: info.Name,
								Description: info.Description,
								Command:     info.Main,
								//Pidfile:
								Path:    filepath.Join(path, file.Name()),
								Respawn: info.Respawn,
								Delay:   info.Delay,
								Ping:    info.Ping,
								Pidfile: Pidfile(file.Name() + ".pid"),
								Logfile: file.Name() + ".log",
								Errfile: file.Name() + ".log",
							}

							c.Processes[file.Name()] = process
						}

					}
				}
			}
		}
	}
	//spew.Dump(c.Processes)
}

func LoadConfig() (*Config, error) {
	return &Config{
		Port:      nconfig.MustInt("director", "port"),
		Username:  nconfig.MustString("director", "username"),
		Password:  nconfig.MustString("director", "password"),
		Pidfile:   Pidfile(nconfig.MustString("director", "pidfile")),
		Logfile:   nconfig.MustString("director", "logfile"),
		Errfile:   nconfig.MustString("director", "errfile"),
		Paths:     nconfig.MustStringArray("director", "paths"),
		Processes: make(map[string]*Process),
	}, nil
}
