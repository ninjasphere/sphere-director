// goforever - processes management
// Copyright (c) 2013 Garrett Woodworth (https://github.com/gwoo).

// sphere-director - Ninja processes management
// Copyright (c) 2014 Ninja Blocks Inc. (https://github.com/ninjablocks).

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/gwoo/greq"
	"github.com/ninjasphere/go-ninja/logger"
)

var config *Config
var daemon *Process
var log = logger.GetLogger("Director")

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	usage := `
Commands
  list              List processes.
  show [name]       Show main proccess or named process.
  start [name]      Start main proccess or named process.
  stop [name]       Stop main proccess or named process.
  restart [name]    Restart main proccess or named process.
`
	fmt.Fprintln(os.Stderr, usage)
}

func init() {
	flag.Usage = Usage
	flag.Parse()
	setConfig()
	daemon = &Process{
		Name:    "director",
		Args:    []string{},
		Command: "director",
		Pidfile: Pidfile(homeDir + "director.pid"),
		Logfile: homeDir + "director.log",
		Errfile: homeDir + "director.log",
		Respawn: 1,
	}

	if runtime.GOOS == "linux" {
		daemon.Path = "/opt/ninjablocks/bin"
	}
}

func main() {

	if len(flag.Args()) > 0 {
		fmt.Printf("%s", Cli())
		return
	}
	if len(flag.Args()) == 0 {
		RunDaemon()
		Mqtt()
		HttpServer()
	}
}

func Cli() string {
	var o []byte
	var err error
	sub := flag.Arg(0)
	name := flag.Arg(1)
	req := greq.New(host(), true)
	if sub == "list" {
		o, _, err = req.Get("/")
	}
	if name == "" {
		if sub == "start" {
			daemon.Args = append(daemon.Args, os.Args[2:]...)
			return daemon.start(daemon.Name)
		}
		_, _, err = daemon.find()
		if err != nil {
			return fmt.Sprintf("Error: %s.\n", err)
		}
		if sub == "show" {
			return fmt.Sprintf("%s.\n", daemon.String())
		}
		if sub == "stop" {
			message := daemon.stop()
			return message
		}
		if sub == "restart" {
			ch, message := daemon.restart()
			fmt.Print(message)
			return fmt.Sprintf("%s\n", <-ch)
		}
	}
	if name != "" {
		path := fmt.Sprintf("/%s", name)
		switch sub {
		case "show":
			o, _, err = req.Get(path)
		case "start":
			o, _, err = req.Post(path, nil)
		case "stop":
			o, _, err = req.Delete(path)
		case "restart":
			o, _, err = req.Put(path, nil)
		}
	}
	if err != nil {
		fmt.Printf("Process error: %s", err)
	}
	return fmt.Sprintf("%s\n", o)
}

func RunDaemon() {
	daemon.children = config.Processes
	//daemon.run()
}

func setConfig() {
	var err error
	config, err = LoadConfig()
	if err != nil {
		log.Fatalf("%s", err)
		return
	}
	config.FindProcesses()
}

func host() string {
	scheme := "https"
	if isHttps() == false {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s:%s@:%d",
		scheme, config.Username, config.Password, config.Port,
	)
}
