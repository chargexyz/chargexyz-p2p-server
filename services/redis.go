package services

import (
	"context"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

type redisServer struct {
	ctx        context.Context
	address    string
	pubChannel string
	subChannel string
	subConn    *redis.PubSubConn
	done       chan error
}

func NewRedisServer(host, port, pubChannel, subChannel string) *redisServer {
	redisServerAddr := fmt.Sprintf("%s:%s", host, port)
	return &redisServer{nil, redisServerAddr, pubChannel, subChannel, nil, nil}
}

func (rs *redisServer) Run(ctx context.Context) error {

	fmt.Println("...Connecting to redis on ", rs.address)
	conn, err := rs.dial()

	if err != nil {
		return err
	}
	rs.ctx = ctx

	rs.subscribe(conn)
	fmt.Println("Connected to redis!")

	return nil
}

func (rs *redisServer) subscribe(cn redis.Conn) {
	rs.done = make(chan error, 1)
	conn := redis.PubSubConn{Conn: cn}

	if err := conn.Subscribe(redis.Args{}.AddFlat(rs.subChannel)...); err != nil {
		rs.done <- err
	}
	rs.subConn = &conn
	// Start listening to redis channel messages
	go rs.listenEvents()

}

func (rs *redisServer) listenEvents() {

	for {
		switch n := rs.subConn.Receive().(type) {
		case error:
			fmt.Println("error message received from redis: ", n.Error())
			// rs.done <- n
			return
		case redis.Message:
			fmt.Println("message received from redis: ", n.Data)

		case redis.Subscription:
			fmt.Printf("Subcribed to Redis %s channel!\n", n.Channel)

		}
	}
}

func (rs *redisServer) Unsubscribe() error {
	if err := rs.subConn.Unsubscribe(); err != nil {
		return err
	}
	fmt.Println("Redis sub channel unsubscribed!")

	// close connection
	defer func() {
		rs.subConn.Conn.Close()
		fmt.Println("Redis connection closed!")
	}()

	return nil
}

func (rs *redisServer) Publish(msg string) error {
	fmt.Printf("...publishing message to redis %s channel on %s\n", rs.pubChannel, rs.address)
	conn, err := rs.dial()

	if err != nil {
		return err
	}
	_, err = conn.Do("PUBLISH", rs.pubChannel, msg)
	if err != nil {
		return err
	}
	fmt.Println("Message published!")
	return nil
}

func (rs *redisServer) dial() (redis.Conn, error) {

	conn, err := redis.Dial("tcp", rs.address)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
