# Peaq Network EV Charging Simulator P2P

A Peer-to-Peer(p2p) application layer that enables the connection and communication between the charging simulator (provider) and the charmEV DApp (consumer). Event messages from the simulator are sent to the p2p and then forwarded to the connected peers requesting the charging service. The p2p server also published event messages from the consumer peer to the simulator (provider peer) for processing.

Provider has to be authenticated before service is requested by the consumer. This is done by the consumer sending a plain hex data to the p2p server. The plain data payload is signed by the p2p server using the provider's private key  and returns the signature to the consumer. Consumer verifies the returned signature using the provider public key.

This application uses the [libp2p gossipsub protocol](https://docs.libp2p.io/concepts/publish-subscribe/#gossip)

The P2P server communicate to the simulator using Redis `PUB/SUB` protocol. The server subscribes to the `OUT` channel where its fetch the event messages from the simulator. It publishes to the `IN` channel where its send all event messages from consumer peer to simulator (provider peer).

### Requirements
* [Download](https://redis.io/download/) and install `Redis`
* Run redis server `redis-server -p 6379`

### Build

Use the following command to build.

`go build -v -a -installsuffix cgo -o out/sim-be-p2p ./cmd`

Or using the docker build command:

`docker build .`

### Run
Make sure redis is running before running the program.

You can run the program without building it by running the command below.

`go run ./cmd -p [P2P_PORT] -sk [PROVIDER_PRIVATE_KEY]`

Or after running `go build` command above

`./out/sim-be-p2p -p [P2P_PORT] -sk [PROVIDER_PRIVATE_KEY]`

Or after running `docker build` command above

`docker run -it --rm --network=host out/sim-be-p2p -p [P2P_PORT] -sk [PROVIDER_PRIVATE_KEY]`

Or using the shell script

`./sim-be-p2p.sh`

The following snippet is displayed on the terminal if server runs successfully.

```
    ...Connecting to redis on  127.0.0.1:6379
    Connected to redis!
    Local Peer ID 12D3KooWCazx4ZLTdrA1yeTTmCy5sGW32SFejztJTGdSZwnGf5Yo
    Listening on... [/ip4/127.0.0.1/tcp/10333]
    Subcribed to Redis OUT channel!
```

The `Local Peer ID` is used to create the p2p connection URL needed to connect to the server. e.g `/ip4/127.0.0.1/tcp/10333/p2p/12D3KooWCazx4ZLTdrA1yeTTmCy5sGW32SFejztJTGdSZwnGf5Yo`


## License

[Apache 2.0](https://choosealicense.com/licenses/apache-2.0/)

