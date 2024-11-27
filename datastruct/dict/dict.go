package dict

import (
	"sync"
	"sync/atomic"
)

type ConcurrentDict struct {
	shards []*shard
	mask   uint32
	count  atomic.Int32 // 元素个数
}

type shard struct {
	m  map[string]interface{}
	mu sync.RWMutex
}
