package engine

import (
	"github.com/issimo1/redis-issimo/abs"
	"github.com/issimo1/redis-issimo/datastruct/dict"
	"github.com/issimo1/redis-issimo/redis/protocol"
	"github.com/issimo1/redis-issimo/utils/timewheel"
	"strings"
)

const (
	dataDictSize = 1 << 16
	ttlDictSize  = 1 << 10
)

type DB struct {
	index      int
	dataDict   *dict.ConcurrentDict
	ttlDict    *dict.ConcurrentDict
	versionMap *dict.ConcurrentDict
	writeAof   func(redisCommand [][]byte)
	delay      *timewheel.Delay
}

func newDB(delay *timewheel.Delay) *DB {
	return &DB{
		dataDict:   dict.NewConcurrent(dataDictSize),
		ttlDict:    dict.NewConcurrent(ttlDictSize),
		versionMap: dict.NewConcurrent(dataDictSize),
		writeAof:   func(redisCommand [][]byte) {},
		delay:      delay,
	}
}

func newBasicDB(delay *timewheel.Delay) *DB {
	return &DB{
		dataDict: dict.NewConcurrent(dataDictSize),
		ttlDict:  dict.NewConcurrent(ttlDictSize),
		writeAof: func(redisCommand [][]byte) {},
		delay:    delay,
	}
}

func (db *DB) setIndex(idx int) {
	db.index = idx
}

func (db *DB) Exec(c abs.Connection, redisCommand [][]byte) protocol.Reply {
	command := strings.ToLower(string(redisCommand[0]))
	if command == "multi" {
		if len(redisCommand) != 1 {
			// todo protocol reply
		}
		return nil
	} else if command == "discard" {
		if len(redisCommand) != 1 {
			// todo
		}
		return nil
	} else if command == "watch" {

	} else if command == "unwatch" {
		if len(redisCommand) != 1 {

		}
	} else if command == "exec" {

	}

	if c != nil && c.IsTransaction() {
		return nil
	}
	return db.generalCommand(c, redisCommand)
}

func validateArity(arity int, cmd [][]byte) bool {
	argNum := len(cmd)
	if arity > 0 {
		return arity == argNum
	}
	return -arity == argNum
}

func (db *DB) generalCommand(c abs.Connection, redisCommand [][]byte) protocol.Reply {
	cmd := strings.ToLower(string(redisCommand[0]))
	cmdFunc, ok := commandCenter[cmd]
	if !ok {
		// unknown command
	}
	if !validateArity(cmdFunc.argsNum, redisCommand) {
		// params doesnt match
	}
	keyFunc := cmdFunc.keyFunc
	readKeys, writeKeys := keyFunc(redisCommand[1:])
	db.addVersion(writeKeys...)
	fun := cmdFunc.execFunc
	return fun(db, redisCommand[1:])
}

func (db *DB) addVersion(keys ...string) {
	for _, key := range keys {
		db.versionMap.AddVersion(key, 1)
	}
}

func (db *DB) execWithLock(cmd [][]byte) protocol.Reply {
	cmds := strings.ToLower(string(cmd[0]))
	cmdFunc, ok := commandCenter[cmds]
	if !ok {
		// unknown command
	}
	if !validateArity(cmdFunc.argsNum, cmd) {
		//
	}
	fun := cmdFunc.execFunc
	return fun(db, cmd[1:])
}

func (db *DB) RWLock(readKeys, writeKeys []string) {
	db.dataDict.RWLock(readKeys, writeKeys)
}

func (db *DB) RWUnLock(readKeys, writeKeys []string) {
	db.dataDict.RWUnLock(readKeys, writeKeys)
}

func genExpireKey(key string) string {
	return "expire:" + key
}
