Phore Websockets server
========================

To make this working, you need to add -blocknotify to your phore daemon, like:

```sh
./phored -blocknotify="~/notify.sh %s"
```

In your `notify.sh` file you need to put the following content:

```sh
#!/bin/sh
curl "http://127.0.0.1:8080/notify" -d "$@"
```

This is used to notify the websockets server of new blocks on Phore blockchain.

Running the server
-------------------

```sh
go run *.go
```

To test to see it working
-------

1) Go to http://localhost:8080/ and click on "Subscribe Address"
2) Run the following command:
```sh
curl "http://127.0.0.1:8080/notify" -d "5359ecc29127cb9a1966caf1b10f57aeec631c5fb01f882fbcaafefeaaae3219"
```
