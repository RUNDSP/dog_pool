//
// Redis Connection Wrapper written in GO
//

package dog_pool

import "errors"
import "fmt"
import "strings"
import "time"
import "github.com/fzzy/radix/redis"
import "github.com/alecthomas/log4go"

//
// Connection Wrapper for Redis
//
type RedisConnection struct {
	Url string "Redis URL this factory will connect to"

	Id string "(optional) Identifier for distingushing between redis connections"

	Logger *log4go.Logger "Handle to the logger we are using"

	Timeout time.Duration "Connection Timeout"

	client *redis.Client "Connection to a Redis, may be nil"
}

//
// Lazily make a Redis Connection
//
func makeLazyRedisConnection(url string, id string, timeout time.Duration, logger *log4go.Logger) (*RedisConnection, error) {
	// Create a new factory instance
	p := &RedisConnection{Url: url, Id: id, Logger: logger, Timeout: timeout}

	// Return the factory
	return p, nil
}

//
// Agressively make a Redis Connection
//
func makeAgressiveRedisConnection(url string, id string, timeout time.Duration, logger *log4go.Logger) (*RedisConnection, error) {
	// Create a new factory instance
	p, _ := makeLazyRedisConnection(url, id, timeout, logger)

	// Ping the server
	if err := p.Ping(); nil != err {
		// Close the connection
		p.Close()

		// Return the error
		return nil, err
	}

	// Return the factory
	return p, nil
}

//
// Clone the connection and return a new instance of RedisConnection
//
func (p *RedisConnection) Clone() *RedisConnection {
	connection, _ := makeLazyRedisConnection(p.Url, p.Id, p.Timeout, p.Logger)
	return connection
}

//
//  ========================================
//
// RedisClientInterface -and- redis.Client implementation:
//
//  ========================================
//

//
// Close closes the connection.
//
func (p *RedisConnection) Close() (err error) {
	// Close the connection
	if nil != p.client {
		err = p.client.Close()
	}

	// Set the pointer to nil
	p.client = nil

	// Log the event
	p.Logger.Info("[RedisConnection][Close][%s/%s] --> Closed!", p.Url, p.Id)

	return
}

//
// Cmd calls the given Redis command:
// - Calls Append(...)
// - Returns GetReply()
//
func (p *RedisConnection) Cmd(cmd string, args ...interface{}) *redis.Reply {
	stop_watch := MakeStopWatch(p, p.Logger, strings.Join([]string{"Cmd", cmd}, " ")).Start()
	defer stop_watch.LogDurationAt(log4go.TRACE)
	defer stop_watch.Stop()

	p.Append(cmd, args...)
	return p.GetReply()
}

//
// Append adds the given call to the pipeline queue.
// Use GetReply() to read the reply.
//
func (p *RedisConnection) Append(cmd string, args ...interface{}) {
	args_to_s := func() string {
		return fmt.Sprintf(strings.Repeat("%v ", len(args)), args...)
	}

	// Wrap in a lambda to prevent evaulation, unless logging is enabled ...
	p.Logger.Trace(func() string {
		args_s := args_to_s()
		return fmt.Sprintf("[RedisConnection][Append][%s/%s] Redis Command = '%s %s'", p.Url, p.Id, cmd, args_s)
	})

	// If the connection is not open, then open it
	if !p.IsOpen() {
		// Did opening the connection fail?
		if err := p.Open(); nil != err {
			p.Logger.Warn(func() string {
				args_s := args_to_s()
				return fmt.Sprintf("[RedisConnection][Append][%s/%s] Redis Command = '%s %s' --> Error = %v", p.Url, p.Id, cmd, args_s, err)
			})
			return
		}
	}

	// Append the command
	stop_watch := MakeStopWatchTags(p, p.Logger, []string{p.Url, p.Id, "Append", cmd}).Start()
	p.client.Append(cmd, args...)
	stop_watch.Stop().LogDurationAt(log4go.FINEST)
}

//
// GetReply returns the reply for the next request in the pipeline queue.
// Error reply with PipelineQueueEmptyError is returned,
// if the pipeline queue is empty.
//
func (p *RedisConnection) GetReply() *redis.Reply {
	// Connection is closed?
	if !p.IsOpen() {
		return &redis.Reply{Type: redis.ErrorReply, Err: ErrConnectionIsClosed}
	}

	// Get the reply from redis
	stop_watch := MakeStopWatchTags(p, p.Logger, []string{p.Url, p.Id, "GetReply"}).Start()
	reply := p.client.GetReply()
	stop_watch.Stop().LogDurationAt(log4go.FINEST)

	// If the connection
	if reply.Type == redis.ErrorReply {
		//* Common errors
		switch reply.Err.Error() {
		case redis.AuthError.Error():
			fallthrough
		case redis.LoadingError.Error():
			fallthrough
		case redis.ParseError.Error():
			fallthrough
		case redis.PipelineQueueEmptyError.Error():
			// Log the error & break
			p.Logger.Warn("[RedisConnection][GetReply][%s/%s] Ignored Error from Redis, Error = %v", p.Url, p.Id, reply.Err)
			break

		default:
			// All other errors are fatal!
			// Close the connection and log the error
			p.Logger.Error("[RedisConnection][GetReply][%s/%s] Fatal Error from Redis, Error = %v", p.Url, p.Id, reply.Err)
			p.Close()
		}
	} else {
		// Log the response
		p.Logger.Trace("[RedisConnection][GetReply][%s/%s] Redis Reply Type = %d, Value = %v", p.Url, p.Id, reply.Type, reply.String())
	}

	// Return the reply from redis to the caller
	return reply
}

func (p *RedisConnection) BatchCommands(commands []*RedisBatchCommand) error {
	stop_watch := MakeStopWatchTags(p, p.Logger, []string{p.Url, p.Id, "BatchCommands"}).Start()

	stop_watch_commands := make([]*StopWatch, len(commands))
	for index, command := range commands {
		stop_watch_commands[index] = MakeStopWatch(p, p.Logger, strings.Join([]string{"BatchCommands", "Cmd", command.Cmd}, " ")).Start()

		if nil == command.Args {
			p.Append(command.Cmd)
		} else {
			args := make([]interface{}, len(command.Args))
			for i, arg := range command.Args {
				args[i] = arg
			}
			p.Append(command.Cmd, args...)
		}
	}

	for index := range commands {
		command := commands[index]
		command.Reply = p.GetReply()

		stop_watch_commands[index].Stop().LogDurationAt(log4go.FINEST)

		if p.IsClosed() {
			return errors.New(fmt.Sprintf("[BatchCommands] Connection closed while getting reply for cmd = %v", command))
		}
	}

	stop_watch.Stop().LogDurationAt(log4go.TRACE)

	return nil
}

//
//  ========================================
//
// RedisConnection implementation:
//
//  ========================================
//

//
// [Depricated, use Append/GetReply above instead]
//
// Get a connection to Redis
//
func (p *RedisConnection) Client() (*redis.Client, error) {
	// Is a saved connection available?
	if p.IsOpen() {
		p.Logger.Trace("[RedisConnection][Client][%s/%s] --> Found Opened Connection!", p.Url, p.Id)

		// Return the connection
		return p.client, nil
	} else {
		p.Logger.Warn("[RedisConnection][Client][%s/%s] --> Found Closed Connection!", p.Url, p.Id)
	}

	// Open a new connection to redis
	if err := p.Open(); nil != err {
		// Abort on errors
		return nil, err
	}

	// Return the new redis connection
	return p.client, nil
}

//
// Ping the server, opening the client connection if necessary
// Returns:
//   nil   --> Ping was successful!
//   error --> Ping was failure
//
func (p *RedisConnection) Ping() error {
	return p.Cmd("ping").Err
}

//
// Return true if the client connection exists
//
func (p *RedisConnection) IsOpen() bool {
	output := nil != p.client

	// Debug logging
	p.Logger.Trace("[RedisConnection][IsOpen][%s/%s] --> %v", p.Url, p.Id, output)

	return output
}

//
// Return true if the client connection exists
//
func (p *RedisConnection) IsClosed() bool {
	output := nil == p.client

	// Debug logging
	p.Logger.Trace("[RedisConnection][IsClosed][%s/%s] --> %v", p.Url, p.Id, output)

	return output
}

//
// Open a new connection to redis
//
func (p *RedisConnection) Open() error {
	// Set the default timeout
	if time.Duration(0) == p.Timeout {
		p.Timeout = time.Duration(10) * time.Second
	}

	// Open the TCP connection
	client, err := redis.DialTimeout("tcp", p.Url, p.Timeout)

	// Check for errors
	if nil != err {
		// Log the event
		p.Logger.Error("[RedisConnection][Open][%s/%s] --> Error = %v", p.Url, p.Id, err)

		// Return the error
		return err
	}

	// Save the client pointer
	p.client = client

	// Log the event
	p.Logger.Info("[RedisConnection][Open][%s/%s] --> Opened!", p.Url, p.Id)

	// Return nil
	return nil
}
