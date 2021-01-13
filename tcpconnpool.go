package TCPconnPool

import (
	"errors"
	"log"
	"net"
	"sync/atomic"
)

type ConnectionPool struct {
	connections chan net.Conn
	maxSize     uint64
}

var TotalConnections uint64

func CreateConnectionPool(initialSize int, maximumSize uint64) (*ConnectionPool, error) {
	log.Println("Connection pool is being created!")
	pool := &ConnectionPool{
		connections: make(chan net.Conn, maximumSize),
		maxSize:     maximumSize,
	}
	TotalConnections = 0

	// Creating the number of initial connections
	for iterator := 0; iterator < initialSize; iterator++ {
		atomic.AddUint64(&TotalConnections, 1)
		singleConnection, er := net.Dial("tcp", "localhost:8081")
		if er != nil {
			log.Fatal("error in creating initial connections: ", er.Error())
		}
		pool.connections <- singleConnection
	}
	return pool, nil
}

func (pool *ConnectionPool) GetOneConnection() (net.Conn, error) {
	if atomic.LoadUint64(&TotalConnections) >= pool.maxSize {
		singleConnection := <-pool.connections
		return singleConnection, nil
	} else {
		select {
		case singleConnection := <-pool.connections:
			if singleConnection == nil {
				return nil, errors.New("returned a nil connection.")
			}
			return singleConnection, nil
		default:
			atomic.AddUint64(&TotalConnections, 1)
			return net.Dial("tcp", "localhost:8081")
		}
	}
}

func (pool *ConnectionPool) PutOneConnection(singleConnection net.Conn) error {
	if singleConnection == nil {
		return nil
	}

	if pool.connections == nil {
		singleConnection.Close()
		return errors.New("pool was already closed")
	}

	select {
	case pool.connections <- singleConnection:
		return nil
	default:
		singleConnection.Close()
		return nil
	}
}
