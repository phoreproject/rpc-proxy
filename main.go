// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/phoreproject/btcd/rpcclient"
	"github.com/phoreproject/rpc-proxy/websockets"
)

// AllowedCommands should not carry state, require sensitive information, or send sensitive data.
var AllowedCommands = map[string]struct{}{
	"getbestblockhash":   struct{}{},
	"getblock":           struct{}{},
	"getblockchaininfo":  struct{}{},
	"getblockcount":      struct{}{},
	"getblockhash":       struct{}{},
	"getblockheader":     struct{}{},
	"getchaintips":       struct{}{},
	"getdifficulty":      struct{}{},
	"getmempoolinfo":     struct{}{},
	"getrawmempool":      struct{}{},
	"gettxout":           struct{}{},
	"gettxoutsetinfo":    struct{}{},
	"getinfo":            struct{}{},
	"getmininginfo":      struct{}{},
	"getnetworkhashps":   struct{}{},
	"submitblock":        struct{}{},
	"getconnectioncount": struct{}{},
	"ping":               struct{}{},
	"masternodelist":     struct{}{},
	"getrawtransaction":  struct{}{},
	"sendrawtransaction": struct{}{},
	"estimatefee":        struct{}{},
	"estimatepriority":   struct{}{},
}

var callbackaddr = flag.String("callbackaddr", ":8080", "http callback address")
var pubaddr = flag.String("pubaddr", ":8000", "public http address")
var host = flag.String("host", "127.0.0.1:11772", "the host and port to connect to phored")
var tls = flag.Bool("disabletls", true, "whether to use SSL/TLS to connect to phored")
var pass = flag.String("pass", "phorepass", "the password to connect to phored")
var user = flag.String("user", "phorerpc", "the username to connect to phored")

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "public/index.html")
}

func main() {
	flag.Parse()

	publicServer := http.NewServeMux()

	connCfg := &rpcclient.ConnConfig{
		Host:                 *host,
		HTTPPostMode:         true,
		User:                 *user,
		Pass:                 *pass,
		DisableTLS:           *tls,
		DisableAutoReconnect: false,
	}

	client, _ := rpcclient.New(connCfg, nil)

	hub := websockets.NewHub()

	notificationServer := websockets.Run(hub, client)

	httpClient := &http.Client{}

	publicServer.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(404)
			w.Write([]byte("not found"))
			return
		}
		type RPCRequest struct {
			JSONRPC string   `json:"jsonrpc"`
			Method  string   `json:"method"`
			Params  []string `json:"params"`
			ID      int      `json:"id"`
		}
		s := make([]byte, r.ContentLength)
		_, err := io.ReadFull(r.Body, s)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("could not read message"))
			return
		}
		request := RPCRequest{}
		err = json.Unmarshal(s, &request)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("could not parse json"))
			fmt.Println(err)
			return
		}

		if _, ok := AllowedCommands[request.Method]; !ok {
			w.WriteHeader(403)
			w.Write([]byte("cannot use command " + request.Method))
			return
		}

		req, err := http.NewRequest("POST", "http://"+*host+"/", bytes.NewReader(s))
		req.SetBasicAuth(*user, *pass)
		res, err := httpClient.Do(req)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("unable to retrieve command output"))
		}

		response := make([]byte, res.ContentLength)
		io.ReadFull(res.Body, response)

		w.WriteHeader(200)
		w.Write(response)
	})

	publicServer.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.ServeWs(hub, w, r, client)
	})

	publicServer.HandleFunc("/", serveHome)

	// Start the callback listening server
	go func() {
		log.Println("Starting callback listener...")
		err := http.ListenAndServe(*callbackaddr, notificationServer)
		if err != nil {
			log.Fatal("error starting callback listener: ", err)
		}
		log.Println("Started callback listener on: ", callbackaddr)
	}()

	log.Println("Starting public listener...")
	err := http.ListenAndServe(*pubaddr, publicServer)
	if err != nil {
		log.Fatal("error starting public listener: ", err)
	}
	log.Println("Started public listener on: ", pubaddr)
}
