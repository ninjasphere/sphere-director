// goforever - processes management
// Copyright (c) 2013 Garrett Woodworth (https://github.com/gwoo).

// sphere-director - Ninja processes management
// Copyright (c) 2014 Ninja Blocks Inc. (https://github.com/ninjablocks).

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HttpServer() {
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/", AuthHandler(Handler))
	fmt.Printf("Sphere Director - Serving port %d\n", config.Port)
	if isHttps() == false {
		http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
		return
	}
	log.Infof("SSL enabled.\n")
	http.ListenAndServeTLS(fmt.Sprintf(":%d", config.Port), "cert.pem", "key.pem", nil)
}

func isHttps() bool {
	_, cerr := os.Open("cert.pem")
	_, kerr := os.Open("key.pem")

	if os.IsNotExist(cerr) || os.IsNotExist(kerr) {
		return false
	}
	return true
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		DeleteHandler(w, r)
		return
	case "POST":
		PostHandler(w, r)
		return
	case "PUT":
		PutHandler(w, r)
		return
	case "GET":
		GetHandler(w, r)
		return
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	var output []byte
	var err error
	switch r.URL.Path[1:] {
	case "":
		output, err = json.Marshal(daemon.children.keys())
	default:
		output, err = json.Marshal(daemon.children.get(r.URL.Path[1:]))
	}
	if err != nil {
		log.Warningf("Get Error: %#v", err)
		return
	}
	fmt.Fprintf(w, "%s", output)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	p := daemon.children.get(name)
	if p == nil {
		fmt.Fprintf(w, "%s does not exist.", name)
		return
	}
	cp, _, _ := p.find()
	if cp != nil {
		fmt.Fprintf(w, "%s already running.", name)
		return
	}
	ch := RunProcess(name, p)
	fmt.Fprintf(w, "%s", <-ch)
}

func PutHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	p := daemon.children.get(name)
	if p == nil {
		fmt.Fprintf(w, "%s does not exist.", name)
		return
	}
	p.find()
	ch, _ := p.restart()
	fmt.Fprintf(w, "%s", <-ch)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	p := daemon.children.get(name)
	if p == nil {
		fmt.Fprintf(w, "%s does not exist.", name)
		return
	}
	p.find()
	p.stop()
	fmt.Fprintf(w, "%s stopped.", name)
}

func AuthHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		for k, v := range r.Header {
			fmt.Printf("  %s = %s\n", k, v[0])
		}
		auth, ok := r.Header["Authorization"]
		if !ok {
			log.Warningf("Unauthorized access to %s", url)
			w.Header().Add("WWW-Authenticate", "basic realm=\"host\"")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Not Authorized.")
			return
		}
		encoded := strings.Split(auth[0], " ")
		if len(encoded) != 2 || encoded[0] != "Basic" {
			log.Warningf("Strange Authorization %q", auth)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded[1])
		if err != nil {
			log.Warningf("Cannot decode %q: %s", auth, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		parts := strings.Split(string(decoded), ":")
		if len(parts) != 2 {
			log.Warningf("Unknown format for credentials %q", decoded)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if parts[0] == config.Username && parts[1] == config.Password {
			fn(w, r)
			return
		}
		log.Warningf("Unauthorized access to %s", url)
		w.Header().Add("WWW-Authenticate", "basic realm=\"host\"")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Not Authorized.")
		return
	}
}
