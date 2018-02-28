# Phore RPC Proxy

This proxy to the Phore network allows easy access to blockchain data for third-party developers without needing to run a Phore node. Any wallets implementing this proxy will need to be implemented trustlessly as all blockchain data allowed by this proxy is stateless and trustless.

### Usage

This proxy can be used by sending JSON-RPC to `/rpc` commands as shown:

```json
{"jsonrpc": "2.0", "method": "getbestblockhash", "params": [], "id": 1}
```

You can read more about JSON-RPC [here](http://www.jsonrpc.org/specification).

### Allowed Methods

Only a subset of the Phore RPC commands are allowed to be used to prevent developers from accessing or sharing sensitive information. These commands can be found in the list below:

- getbestblockhash
- getblock
- getblockchaininfo
- getblockcount
- getblockhash
- getblockheader
- getchaintips
- getdifficulty
- getmempoolinfo
- getrawmempool
- gettxout
- gettxoutsetinfo
- getinfo
- getmininginfo
- getnetworkhashps
- submitblock
- getconnectioncount
- ping
- masternodelist
- getrawtransaction
- sendrawtransaction
- estimatefee
- estimatepriority

If you try to access a method not in this list, the proxy will return:
```json
{"err": "Forbidden to run command \"badcommand\""}
```

## Websockets

Websockets are also supported and allow for notifications based on new blocks, transactions, and bloom filters.

#### `subscribeBloom <bloomFilter> [includeMempool=false]`
This command will subscribe to any transactions that are included in the bloom filter provided.

#### `subscribeAddress <address> [includeMempool=false]`
This command will subscribe to a certain address.

#### `subscribeBlock`
This command will subscribe to any blocks that are staked.

#### `unsubscribeAll`
This command will cancel any subscriptions currently active.