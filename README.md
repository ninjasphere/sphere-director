# Sphere Directory [![Build Status](https://travis-ci.org/ninjasphere/sphere-director.png)](https://travis-ci.org/ninjasphere/sphere-director)

Ninja Sphere's process manager. Used to start, stop, restart and monitor Sphere processes (like drivers).

Based *heavily* on [goforever](https://github.com/gwoo/goforever) by [Garrett Woodworth]((https://github.com/gwoo)

## Running
Help.

	./director -h

Daemonize main process.

	./director start

Run main process and output to current session.

	./director

## CLI
	list				List processes.
	show [process]	    Show a main proccess or named process.
	start [process]		Start a main proccess or named process.
	stop [process]		Stop a main proccess or named process.
	restart [process]	Restart a main proccess or named process.

## HTTP API

Return a list of managed processes

	GET host:port/

Start the process

	POST host:port/:name

Restart the process

	PUT host:port/:name

Stop the process

	DELETE host:port/:name
