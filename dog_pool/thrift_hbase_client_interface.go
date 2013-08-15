//
// Thrift+Hbase Client Interface
//
// Interface implemented by thrift.Hbase and dog_pool.ThriftConnection
//

package dog_pool

import "./thrift"

type ThriftClientInterface interface {
	// Implemenent all of the client methods
	thrift.Hbase

	// Plus these methods too ...

	// Is the connection open?
	IsOpen() bool
	// Is the connection closed?
	IsClosed() bool

	// Open the connection, return error on failure
	Open(url string) error

	// Close the connection
	Close() error
}
