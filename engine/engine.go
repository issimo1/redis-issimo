package engine

import (
	"github.com/issimo1/redis-issimo/abs"
	"github.com/issimo1/redis-issimo/engine/payload"
	"github.com/issimo1/redis-issimo/redis/protocol"
	"github.com/issimo1/redis-issimo/utils/config"
	"github.com/issimo1/redis-issimo/utils/logger"
	"github.com/issimo1/redis-issimo/utils/timewheel"
	"strings"
	"sync/atomic"
	"time"
)

type Engine struct {
	dbSet []*atomic.Value
	delay *timewheel.Delay
}

func NewEngine() *Engine {
	engine := &Engine{}

	engine.dbSet = make([]*atomic.Value, config.GlobalConfig.DBCount)
	for idx := 0; idx < config.GlobalConfig.DBCount; idx++ {
		db := newDB(engine.delay)
		db.setIndex(idx)
		dbSet := &atomic.Value{}
		dbSet.Store(db)
		engine.dbSet[idx] = dbSet
	}

	if config.GlobalConfig.AppendOnly {

	}
	return engine
}

func (e *Engine) aofBindEveryDB() {
	for _, dbset := range e.dbSet {
		db := dbset.Load().(*DB)
		db.writeAof = func(redisCommand [][]byte) {
			if config.GlobalConfig.AppendOnly {

			}
		}
	}
}

func (e *Engine) selectDB(idx int) (*DB, protocol.Reply) {
	if idx < 0 || idx >= len(e.dbSet) {
		return nil, nil
	}
	return e.dbSet[idx].Load().(*DB), nil
}

func (e *Engine) Exec(c abs.Connection, redisCommand [][]byte) protocol.Reply {
	defer func() {
		if err := recover(); err != nil {

		}
	}()
	cmd := strings.ToLower(string(redisCommand[0]))
	if cmd == "ping" {
		return Ping(redisCommand[1:])
	}
	if cmd == "auth" {
		return Auth(c, redisCommand[1:])
	}
	if !checkPassword(c) {
		// err
	}

	switch cmd {
	case "select":
		if c != nil && c.IsTransaction() {
			// can not select db
		}
		return nil
	case "bgrewriteaof":
		if !config.GlobalConfig.AppendOnly {
			// appendonly false
		}
	case "subscribe":
	case "unsubscribe":
	case "publish":
	}
	dbIdx := c.GetDBIdx()
	logger.Debugf("db index:%d", dbIdx)
	db, err := e.selectDB(dbIdx)
	if err != nil {
	}
	return db.Exec(c, redisCommand)
}

func (e *Engine) Close() {

}

func (e *Engine) RWLocks(dbIdx int, readKeys, writeKeys []string) {
	db, err := e.selectDB(dbIdx)
	if err != nil {
		logger.Error("RWLock err:", err)
		return
	}
	db.RWLock(readKeys, writeKeys)
}

func (e *Engine) RWUnLocks(dbIdx int, readKeys, writeKeys []string) {
	db, err := e.selectDB(dbIdx)
	if err != nil {
		// err
		logger.Error("RWUnLock err:", err)
		return
	}
	db.RWUnLock(readKeys, writeKeys)
}

func (e *Engine) GetUndoLogs(dbIdx int, redisCommand [][]byte) {

}

func (e *Engine) ExecWithLock(dbIdx int, redisCommand [][]byte) protocol.Reply {
	db, err := e.selectDB(dbIdx)
	if err != nil {
		logger.Error("RWLocks err:", err)
		return nil
	}
	return db.execWithLock(redisCommand)
}

func (e *Engine) ForEach(dbIdx int, callBack func(key string, data *payload.DataEntity, expiration *time.Time) bool) {
	db, err := e.selectDB(dbIdx)
	if err != nil {
		logger.Error("RWLocks err:", err)
		return
	}

	db.dataDict.ForEach(func(key string, val interface{}) bool {
		entity, _ := val.(*payload.DataEntity)
		var expiration *time.Time
		rawExpireTime, ok := db.ttlDict.Get(key)
		if ok {
			expireTime, _ := rawExpireTime.(time.Time)
			expiration = &expireTime
		}
		return callBack(key, entity, expiration)
	})
}

func newAuxiliaryEngine() *Engine {
	engine := &Engine{}
	engine.delay = timewheel.NewDelay()
	engine.dbSet = make([]*atomic.Value, config.GlobalConfig.DBCount)
	for i := range engine.dbSet {

		db := newBasicDB(engine.delay)
		db.setIndex(i)

		holder := &atomic.Value{}
		holder.Store(db)
		engine.dbSet[i] = holder
	}
	return engine
}
