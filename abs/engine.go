package abs

import (
	"github.com/issimo1/redis-issimo/redis/protocol"
	"time"
)

type Engine interface {
	Exec(Connection, [][]byte) protocol.Reply
	ForEach(int, func(key string, data string, expiration *time.Time) bool)
	Close()
}
