package server

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisServer struct {
	host       string
	port       string
	pubChannel string
	subChannel string
	conn       redis.Conn
}

func NewRedisServer(host, port, pubChannel, subChannel string) *redisServer {
	return &redisServer{host, port, pubChannel, subChannel, nil}
}

func (rs *redisServer) Run() error {
	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	const healthCheckPeriod = time.Minute

	redisServerAddr := fmt.Sprintf("%s:%s", rs.host, rs.port)

	fmt.Println("Connecting to redis...")

	conn, err := redis.Dial("tcp", redisServerAddr,
		// Read timeout on server should be greater than ping period.
		redis.DialReadTimeout(healthCheckPeriod+10*time.Second),
		redis.DialWriteTimeout(10*time.Second))
	if err != nil {
		return err
	}
	rs.conn = conn
	fmt.Println("Connected to redis!")

	// close connection
	defer func() {
		rs.conn.Close()
		fmt.Println("Redis connection closed!")
	}()

	return nil
}
