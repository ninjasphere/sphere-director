// goforever - processes management
// Copyright (c) 2013 Garrett Woodworth (https://github.com/gwoo).

// sphere-director - Ninja processes management
// Copyright (c) 2014 Ninja Blocks Inc. (https://github.com/ninjablocks).

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var defaultPingTime = "1m"

//Run the process
func RunProcess(p *Process) chan *Process {
	ch := make(chan *Process)
	go func() {
		p.start()

		pid := p.Pid
		p.ping(func(time time.Duration, p *Process) {
			if pid == p.Pid {
				p.Restarts = 0
				fmt.Printf("%s refreshed after %s.\n", p.ID, time)
				p.Status = "running"
			}
		})
		go p.watch()
		ch <- p
	}()
	return ch
}

type Process struct {
	ID   string // The id used to refer to this process... i.e. director start driver-zigbee
	Info packageJson

	Path    string // Optional path for the command to be run in (cwd)
	Command string
	Args    []string // Optional args

	Pidfile Pidfile
	Logfile string
	Errfile string

	PingTime  string // Amount of time to indicate a successful start
	MaxMemory int    // Maximum memory in kilobytes

	FirstStart  time.Time
	LastStart   time.Time
	Restarts    int // Number of restarts since the last successful start, or the beginning
	TotalStarts int // Number of starts since the beginning
	Pid         int
	Status      string

	x *os.Process

	children children
}

func (p *Process) String() string {
	js, err := json.Marshal(p)
	if err != nil {
		log.Warningf("%s", err)
		return ""
	}
	return string(js)
}

//Find a process by name
func (p *Process) find() (*os.Process, string, error) {
	if p.Pidfile == "" {
		return nil, "", errors.New("Pidfile is empty.")
	}
	if pid := p.Pidfile.read(); pid > 0 {
		process, err := os.FindProcess(pid)
		if err != nil {
			return nil, "", err
		}
		p.x = process
		p.Pid = process.Pid
		p.Status = "running"
		message := fmt.Sprintf("%s is %#v\n", p.ID, process.Pid)
		return process, message, nil
	}
	message := fmt.Sprintf("%s not running.\n", p.ID)
	return nil, message, errors.New(fmt.Sprintf("Could not find process %s.", p.ID))
}

//Start the process
func (p *Process) start() string {
	wd := p.Path
	if wd == "" {
		wd, _ = os.Getwd()
	}
	proc := &os.ProcAttr{
		Dir: wd,
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
		},
	}

	if p.Logfile != "" {
		proc.Files = append(proc.Files, NewLog(p.Logfile))
	}
	if p.Errfile != "" {
		proc.Files = append(proc.Files, NewLog(p.Errfile))
	}

	args := append([]string{p.ID}, p.Args...)
	process, err := os.StartProcess(p.Command, args, proc)
	if err != nil {
		log.Warningf("%s failed. %s\n", p.ID, err)
		p.Status = fmt.Sprintf("Failed: %s", err)
		return ""
	}
	err = p.Pidfile.write(process.Pid)
	if err != nil {
		log.Warningf("%s pidfile error: %s\n", p.ID, err)
		p.Status = fmt.Sprintf("Failed to write pid: %s", err)
		return ""
	}
	p.x = process
	p.Pid = process.Pid
	p.Status = "started"

	p.LastStart = time.Now()
	if p.TotalStarts == 0 {
		p.FirstStart = time.Now()
	}
	p.TotalStarts++

	return fmt.Sprintf("%s is %#v\n", p.ID, process.Pid)
}

//Stop the process
func (p *Process) stop() string {
	if p.x != nil {
		// p.x.Kill() this seems to cause trouble
		cmd := exec.Command("kill", fmt.Sprintf("%d", p.x.Pid))
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Warningf("%s", err)
		}
		p.children.stop("all")
	}
	p.release("stopped")
	message := fmt.Sprintf("%s stopped.\n", p.ID)
	return message
}

//Release process and remove pidfile
func (p *Process) release(status string) {
	if p.x != nil {
		p.x.Release()
	}
	p.Pid = 0
	p.Pidfile.delete()
	p.Status = status
}

//Restart the process
func (p *Process) restart() (chan *Process, string) {
	p.stop()
	message := fmt.Sprintf("%s restarted.\n", p.ID)
	ch := RunProcess(p)
	return ch, message
}

//Run callback on the process after given duration.
func (p *Process) ping(f func(t time.Duration, p *Process)) {
	t, err := time.ParseDuration(p.PingTime)
	if err != nil {
		t, _ = time.ParseDuration(defaultPingTime)
	}
	go func() {
		select {
		case <-time.After(t):
			f(t, p)
		}
	}()
}

//Watch the process
func (p *Process) watch() {
	if p.x == nil {
		p.release("stopped")
		return
	}
	status := make(chan *os.ProcessState)
	died := make(chan error)
	go func() {
		state, err := p.x.Wait()
		if err != nil {
			died <- err
			return
		}
		status <- state
	}()
	select {
	case s := <-status:
		if p.Status == "stopped" {
			return
		}
		fmt.Fprintf(os.Stderr, "%s %s\n", p.ID, s)
		fmt.Fprintf(os.Stderr, "%s success = %#v\n", p.ID, s.Success())
		fmt.Fprintf(os.Stderr, "%s exited = %#v\n", p.ID, s.Exited())
		/*if p.Respawn != -1 && p.respawns > p.Respawn {
			p.release("exited")
			log.Warningf("%s respawn limit reached.\n", p.ID)
			return
		}*/
		//fmt.Fprintf(os.Stderr, "%s respawns = %#v\n", p.ID, p.respawns)
		/*if p.Delay != "" {
			t, _ := time.ParseDuration(p.Delay)
			time.Sleep(t)
		}*/
		p.restart()
		p.Status = "restarted"
		p.Restarts++
	case err := <-died:
		p.release("killed")
		log.Infof("%d %s killed = %#v", p.x.Pid, p.ID, err)
	}
}

//Run child processes
func (p *Process) run() {
	for _, p := range p.children {
		RunProcess(p)
	}
}

//Child processes.
type children map[string]*Process

//Stringify
func (c children) String() string {
	js, err := json.Marshal(c)
	if err != nil {
		log.Warningf("%s", err)
		return ""
	}
	return string(js)
}

//Get child processes names.
func (c children) keys() []string {
	keys := []string{}
	for k, _ := range c {
		keys = append(keys, k)
	}
	return keys
}

func (c children) stop(name string) {
	if name == "all" {
		for name, p := range c {
			p.stop()
			delete(c, name)
		}
		return
	}
	p := c[name]
	p.stop()
	delete(c, name)
}

type Pidfile string

//Read the pidfile.
func (f *Pidfile) read() int {
	data, err := ioutil.ReadFile(string(*f))
	if err != nil {
		return 0
	}
	pid, err := strconv.ParseInt(string(data), 0, 32)
	if err != nil {
		return 0
	}
	return int(pid)
}

//Write the pidfile.
func (f *Pidfile) write(data int) error {
	err := ioutil.WriteFile(string(*f), []byte(strconv.Itoa(data)), 0660)
	if err != nil {
		return err
	}
	return nil
}

//Delete the pidfile
func (f *Pidfile) delete() bool {
	_, err := os.Stat(string(*f))
	if err != nil {
		return true
	}
	err = os.Remove(string(*f))
	if err == nil {
		return true
	}
	return false
}

//Create a new file for logging
func NewLog(path string) *os.File {
	if path == "" {
		return nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	return file
}
