// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package websockets handles websocket messages and provides a simple websockets server.
package websockets

import (
	"github.com/phoreproject/btcutil"
	"github.com/phoreproject/btcutil/bloom"
)

// RegisterAddress is a channel used to register an address to a websocket client
type RegisterAddress struct {
	client  *Client
	address string
	mempool IncludeTransaction
}

// RegisterBloom is a channel message used to register a bloom filter
type RegisterBloom struct {
	client  *Client
	bloom   *bloom.Filter
	mempool IncludeTransaction
}

// BroadcastAddressMessage used to receive message of addresses
type BroadcastAddressMessage struct {
	address string
	message []byte
	mempool bool
}

// BroadcastTransactionMessage is used to receive messages of transactions
type BroadcastTransactionMessage struct {
	transaction *btcutil.Tx
	message     []byte
	mempool     bool
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	subscribedToBlocks         map[*Client]bool
	subscribedToAddress        map[string][]*Client
	subscribedToAddressMempool map[string][]*Client
	subscribedToBloom          map[*Client]*bloom.Filter
	subscribedToBloomMempool   map[*Client]*bloom.Filter

	// Output messages to the clients.
	broadcastBlock   chan []byte
	broadcastAddress chan BroadcastAddressMessage
	broadcastBloom   chan BroadcastTransactionMessage

	// Register requests from the clients.
	registerBlock   chan *Client
	registerAddress chan RegisterAddress
	registerBloom   chan RegisterBloom

	// Unregister requests from clients.
	unregister     chan *Client
	unsubscribeAll chan *Client
}

// NewHub creates a new hub to track messages about clients
func NewHub() *Hub {
	return &Hub{
		broadcastBlock:             make(chan []byte),
		broadcastAddress:           make(chan BroadcastAddressMessage),
		broadcastBloom:             make(chan BroadcastTransactionMessage),
		registerBlock:              make(chan *Client),
		registerAddress:            make(chan RegisterAddress),
		registerBloom:              make(chan RegisterBloom),
		unsubscribeAll:             make(chan *Client),
		subscribedToBlocks:         make(map[*Client]bool),
		subscribedToAddress:        make(map[string][]*Client),
		subscribedToAddressMempool: make(map[string][]*Client),
		subscribedToBloom:          make(map[*Client]*bloom.Filter),
		subscribedToBloomMempool:   make(map[*Client]*bloom.Filter),
	}
}

// Run runs the hub forever
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.registerBlock:
			h.subscribedToBlocks[client] = true
		case registerAddress := <-h.registerAddress:
			addr := registerAddress.address
			if registerAddress.mempool == IncludeAllTransactions {
				h.subscribedToAddressMempool[addr] = append(h.subscribedToAddressMempool[addr], registerAddress.client)
				h.subscribedToAddress[addr] = append(h.subscribedToAddress[addr], registerAddress.client)
			} else if registerAddress.mempool == IncludeConfirmedTransactions {
				h.subscribedToAddress[addr] = append(h.subscribedToAddress[addr], registerAddress.client)
			} else {
				h.subscribedToAddressMempool[addr] = append(h.subscribedToAddressMempool[addr], registerAddress.client)
			}
		case registerBloom := <-h.registerBloom:
			if registerBloom.mempool == IncludeAllTransactions {
				h.subscribedToBloomMempool[registerBloom.client] = registerBloom.bloom
				h.subscribedToBloom[registerBloom.client] = registerBloom.bloom
			} else if registerBloom.mempool == IncludeConfirmedTransactions {
				h.subscribedToBloom[registerBloom.client] = registerBloom.bloom
			} else {
				h.subscribedToBloomMempool[registerBloom.client] = registerBloom.bloom
			}
		case client := <-h.unsubscribeAll:
			if _, ok := h.subscribedToBlocks[client]; ok {
				delete(h.subscribedToBlocks, client)
				close(client.send)
			}
		case message := <-h.broadcastBlock:
			for client := range h.subscribedToBlocks {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.subscribedToBlocks, client)
				}
			}
		case broadcastAddress := <-h.broadcastAddress:
			addr := broadcastAddress.address
			var channel []*Client
			if broadcastAddress.mempool {
				channel = h.subscribedToAddressMempool[addr]
			} else {
				channel = h.subscribedToAddress[addr]
			}
			for _, client := range channel {
				go func(c *Client) { // process each message asynchronously
					select {
					case c.send <- broadcastAddress.message:
					default:
						deleteClientFromAddress(c, addr)
						close(c.send)
					}
				}(client)
			}
		case broadcastBloom := <-h.broadcastBloom:
			var channel map[*Client]*bloom.Filter
			if broadcastBloom.mempool {
				channel = h.subscribedToBloomMempool
			} else {
				channel = h.subscribedToBloom
			}
			for client, bloom := range channel {
				if bloom.MatchTxAndUpdate(broadcastBloom.transaction) {
					client.send <- broadcastBloom.message
				}
			}
		case client := <-h.unsubscribeAll:
			delete(h.subscribedToBlocks, client)
			// TODO: Improve this delete method
			for address, clients := range h.subscribedToAddress {
				if clientInSlice(client, clients) {
					deleteClientFromAddress(client, address)
				}
			}
		}
	}
}

func deleteClientFromAddress(client *Client, addr string) {
	var i int
	for j, v := range client.Hub.subscribedToAddress[addr] {
		if v == client {
			i = j
		}
	}
	client.Hub.subscribedToAddress[addr] = append(client.Hub.subscribedToAddress[addr][:i], client.Hub.subscribedToAddress[addr][i+1:]...)
}

func clientInSlice(client *Client, list []*Client) bool {
	for _, b := range list {
		if b == client {
			return true
		}
	}
	return false
}
