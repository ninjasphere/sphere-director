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
	"runtime"

	"github.com/davecgh/go-spew/spew"
	nconfig "github.com/ninjasphere/go-ninja/config"
)

var homeDir = "./"

func init() {
	if runtime.GOOS == "linux" {
		homeDir = "/var/run/director/"
		os.MkdirAll(homeDir, 0777)
	}
}

type Config struct {
	Port      int
	Username  string
	Password  string
	Daemonize bool
	Paths     []string
	Processes map[string]*Process
}

type packageJson struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	Main        string `json:"main,omitempty"`
	Author      string `json:"author,omitempty"`
	License     string `json:"license,omitempty"`
	MaxMemory   int    `json:"maxMemory,omitempty"`
	PingTime    string `json:"ping,omitempty"`
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

						if err != nil || info.Main == "" {
							log.Warningf("Could not read package.info for module %s : %s", file.Name(), err)
						} else {
							//spew.Dump(info)
							process := &Process{
								ID:        file.Name(),
								Info:      info,
								MaxMemory: info.MaxMemory,
								Command:   info.Main,
								Path:      filepath.Join(path, file.Name()),
								PingTime:  info.PingTime,
								Pidfile:   Pidfile(homeDir + file.Name() + ".pid"),
							}

							if process.PingTime == "" {
								process.PingTime = "1m"
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
		Paths:     nconfig.MustStringArray("director", "paths"),
		Processes: make(map[string]*Process),
	}, nil
}
