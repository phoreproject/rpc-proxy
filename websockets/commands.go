package websockets

import (
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/phoreproject/btcd/wire"
	"github.com/phoreproject/btcutil/bloom"
)

// IncludeTransaction is whether to include mempool transactions, confirmed transactions, or all
type IncludeTransaction int

const (
	// IncludeAllTransactions - Include both confirmed and mempool transactions
	IncludeAllTransactions = IncludeTransaction(iota)

	// IncludeMempoolTransactions - Include only mempool transactions
	IncludeMempoolTransactions = IncludeTransaction(iota)

	// IncludeConfirmedTransactions - Include only confirmed transaction
	IncludeConfirmedTransactions = IncludeTransaction(iota)
)

func subscribeBloom(client *Client, args []string) error {
	// Syntax: subscribeBloom <filter> <hashfuncs> <tweak> <includeMempoolTransactions> [flags=UpdateNone]

	if len(args) < 4 || len(args) > 5 {
		return errors.New("Incorrect number of arguments")
	}

	flags := wire.BloomUpdateNone
	if len(args) < 5 {
		flags = wire.BloomUpdateNone
	} else {
		flagsInt, err := strconv.Atoi(args[4])
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

	transactions, err := strconv.Atoi(args[3])
	if err != nil {
		return errors.New("Could not parse includeMempoolTransactions")
	}
	if transactions > 2 || transactions < 0 {
		return errors.New("Could not parse includeMempoolTransactions")
	}

	filter := bloom.LoadFilter(&wire.MsgFilterLoad{
		Filter:    bloomBytes,
		HashFuncs: hashFuncs,
		Tweak:     tweak,
		Flags:     flags,
	})

	client.Hub.registerBloom <- RegisterBloom{client: client, bloom: filter, mempool: IncludeTransaction(transactions)}
	return nil
}

// SubscribeAddress is used for a client to subscribe to any events happening to an address
func subscribeAddress(client *Client, addr string, transactionType string) error {
	transaction, err := strconv.Atoi(transactionType)
	if err != nil {
		return err
	}
	if transaction < 0 || transaction > 3 {
		return errors.New("Incorrect value for includeMempoolTransaction")
	}
	register := RegisterAddress{client: client, address: addr, mempool: IncludeTransaction(transaction)}
	client.Hub.registerAddress <- register
	return nil
}

func subscribeBlock(client *Client) {
	// fmt.Println("One new client registered", client)
	client.Hub.registerBlock <- client
}

func unsubscribeAll(client *Client) {
	client.Hub.unsubscribeAll <- client
}
