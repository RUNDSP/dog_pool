package dog_pool

import "time"

var cmd_bitop = "BITOP"
var cmd_bitop_and = []byte("AND")
var cmd_bitop_not = []byte("NOT")
var cmd_bitop_or = []byte("OR")
var cmd_getbit = "GETBIT"
var cmd_setbit = "SETBIT"
var cmd_bitcount = "BITCOUNT"

var cmd_expire = "EXPIRE"
var cmd_mget = "MGET"
var cmd_get = "GET"
var cmd_del = "DEL"
var cmd_set = "SET"

//
// Factory Methods:
//

// Basic factory method
func MakeRedisBatchCommand(cmd string) *RedisBatchCommand {
	return &RedisBatchCommand{cmd, [][]byte{}, nil}
}

// MGET <KEY> <KEY> <KEY>...
func MakeRedisBatchCommandMget(keys ...string) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_mget,
		args:  make([][]byte, len(keys))[0:0],
		reply: nil,
	}
	output.WriteStringArgs(keys)
	return output
}

// EXPIRE <KEY> <SECONDS>
func MakeRedisBatchCommandExpireIn(key string, expire_in time.Duration) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_expire,
		args:  make([][]byte, 2)[0:0],
		reply: nil,
	}
	output.WriteStringArg(key)
	output.WriteIntArg(int64(expire_in.Seconds()))
	return output
}

// GET <KEY>
func MakeRedisBatchCommandGet(key string) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_get,
		args:  make([][]byte, 1)[0:0],
		reply: nil,
	}
	output.WriteStringArg(key)
	return output
}

// SET <KEY> <VALUE>
func MakeRedisBatchCommandSet(key string, value []byte) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_set,
		args:  make([][]byte, 2)[0:0],
		reply: nil,
	}
	output.WriteStringArg(key)
	output.WriteArg(value)
	return output
}

// DEL <KEY> <KEY> <KEY> ....
func MakeRedisBatchCommandDelete(keys ...string) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_del,
		args:  make([][]byte, len(keys))[0:0],
		reply: nil,
	}
	output.WriteStringArgs(keys)
	return output
}

// BITOP AND <DEST> <SRC KEYS> ...
func MakeRedisBatchCommandBitopAnd(dest string, sources ...string) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_bitop,
		args:  make([][]byte, 2+len(sources))[0:0],
		reply: nil,
	}
	output.WriteArg(cmd_bitop_and)
	output.WriteStringArg(dest)
	output.WriteStringArgs(sources)
	return output
}

// BITOP NOT <DEST> <SRC>
func MakeRedisBatchCommandBitopNot(dest string, source string) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_bitop,
		args:  make([][]byte, 3)[0:0],
		reply: nil,
	}
	output.WriteArg(cmd_bitop_not)
	output.WriteStringArg(dest)
	output.WriteStringArg(source)
	return output
}

// BITOP OR <DEST> <SRC KEYS> ...
func MakeRedisBatchCommandBitopOr(dest string, sources ...string) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_bitop,
		args:  make([][]byte, 2+len(sources))[0:0],
		reply: nil,
	}
	output.WriteArg(cmd_bitop_or)
	output.WriteStringArg(dest)
	output.WriteStringArgs(sources)
	return output
}

// GETBIT <KEY> <INDEX>
func MakeRedisBatchCommandGetBit(key string, index int64) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_getbit,
		args:  make([][]byte, 2)[0:0],
		reply: nil,
	}
	output.WriteStringArg(key)
	output.WriteIntArg(index)
	return output
}

// SETBIT <KEY> <INDEX> <STATE>
func MakeRedisBatchCommandSetBit(key string, index int64, state bool) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_setbit,
		args:  make([][]byte, 3)[0:0],
		reply: nil,
	}
	output.WriteStringArg(key)
	output.WriteIntArg(index)
	output.WriteBoolArg(state)
	return output
}

// BITCOUNT <KEY>
func MakeRedisBatchCommandBitCount(key string) *RedisBatchCommand {
	output := &RedisBatchCommand{
		cmd:   cmd_bitcount,
		args:  make([][]byte, 1)[0:0],
		reply: nil,
	}
	output.WriteStringArg(key)
	return output
}