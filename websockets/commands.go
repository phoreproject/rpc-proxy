package websockets

import (
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/phoreproject/btcd/wire"
	"github.com/phoreproject/btcutil/bloom"
)

func subscribeBloom(client *Client, args []string) error {
	// Syntax: subscribeBloom <filter> <hashfuncs> <tweak> [flags=UpdateNone]

	if len(args) < 3 || len(args) > 4 {
		return errors.New("Incorrect number of arguments")
	}

	flags := wire.BloomUpdateNone
	if len(args) == 3 {
		flags = wire.BloomUpdateNone
	} else {
		flagsInt, err := strconv.Atoi(args[3])
		if err != nil {
			return errors.New("Could not parse update flags")
		}
		flags = wire.BloomUpdateType(flagsInt)
	}

	var bloomBytes []byte
	bloomBytes, err := hex.DecodeString(args[0])
	if err != nil {
		return errors.New("Could not decode bloom filter")
	}

	hashFuncsInt, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("Could not parse HashFuncs")
	}
	hashFuncs := uint32(hashFuncsInt)

	tweakInt, err := strconv.Atoi(args[2])
	if err != nil {
		return errors.New("Could not parse Tweak")
	}
	tweak := uint32(tweakInt)

	filter := bloom.LoadFilter(&wire.MsgFilterLoad{
		Filter:    bloomBytes,
		HashFuncs: hashFuncs,
		Tweak:     tweak,
		Flags:     flags,
	})

	client.Hub.registerBloom <- RegisterBloom{client: client, bloom: filter}
	return nil
}

// SubscribeAddress is used for a client to subscribe to any events happening to an address
func subscribeAddress(client *Client, addr string, mempool bool) {
	// fmt.Println("One new address registered", client, addr)
	register := RegisterAddress{client: client, address: addr, mempool: mempool}
	client.Hub.registerAddress <- register
}

func subscribeBlock(client *Client) {
	// fmt.Println("One new client registered", client)
	client.Hub.registerBlock <- client
}

func unsubscribeAll(client *Client) {
	client.Hub.unsubscribeAll <- client
}
