package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"

	"github.com/phoreproject/btcd/chaincfg"
	"github.com/phoreproject/btcd/wire"
	"github.com/phoreproject/btcutil"
	"github.com/phoreproject/btcutil/bloom"
)

func main() {
	addresses := os.Args[1:]

	f := bloom.NewFilter(uint32(len(addresses)), rand.Uint32(), 0.001, wire.BloomUpdateNone)

	for i := range addresses {
		a, err := btcutil.DecodeAddress(addresses[i], &chaincfg.MainNetParams)
		if err != nil {
			panic(err)
		}
		f.Add(a.ScriptAddress())
	}
	fmt.Println(hex.EncodeToString(f.MsgFilterLoad().Filter))
	fmt.Println(f.MsgFilterLoad().HashFuncs)
	fmt.Println(f.MsgFilterLoad().Tweak)
}
