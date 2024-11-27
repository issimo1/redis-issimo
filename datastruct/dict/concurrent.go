package dict

import (
	"github.com/issimo1/redis-issimo/utils"
	"sort"
	"sync"
	"sync/atomic"
)

type ConcurrentDict struct {
	shards []*shard
	mask   uint32        // 掩码不知道有什么用
	count  *atomic.Int32 // 元素个数
}

type shard struct {
	m  map[string]interface{}
	mu sync.RWMutex
}

func (s *shard) forEach(consumer Consumer) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.m {
		res := consumer(k, v)
		if !res {
			return false
		}
	}
	return true
}

func NewConcurrent(shardCount int) *ConcurrentDict {
	shardCount = utils.ComputeCapacity(shardCount)

	dict := ConcurrentDict{}
	shards := make([]*shard, shardCount)

	for i := range shards {
		shards[i] = &shard{
			m: make(map[string]interface{}),
		}
	}
	dict.shards = shards
	dict.mask = uint32(shardCount - 1)
	dict.count = &atomic.Int32{}
	return &dict
}

func (c *ConcurrentDict) index(code uint32) uint32 {
	return code & c.mask
}

func (c *ConcurrentDict) getShard(key string) *shard {
	return c.shards[c.index(utils.Fnv32(key))]
}

func (c *ConcurrentDict) AddVersion(key string, delta int64) (val interface{}, exist bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	val, exist = shard.m[key]
	if !exist {
		shard.m[key] = delta
		return
	}
	v, ok := val.(int64)
	if ok {
		v = v + delta
	} else {
		v = delta
	}
	shard.m[key] = v
	return v, exist
}

func (c *ConcurrentDict) Get(key string) (val interface{}, exist bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	val, exist = shard.m[key]
	return
}

func (c *ConcurrentDict) GetWithoutLock(key string) (val interface{}, exist bool) {
	if c == nil {
		panic("dict cant be nil")
	}

	shard := c.getShard(key)
	val, exist = shard.m[key]
	return
}

func (c *ConcurrentDict) Count() int {
	return int(c.count.Load())
}

func (c *ConcurrentDict) addCount() {
	c.count.Add(1)
}

func (c *ConcurrentDict) subCount() {
	c.count.Add(-1)
}

func (c *ConcurrentDict) Delete(key string) (interface{}, int) {
	shd := c.getShard(key)
	shd.mu.RLock()
	defer shd.mu.RUnlock()

	if val, ok := shd.m[key]; ok {
		delete(shd.m, key)
		c.subCount()
		return val, 1
	}
	return nil, 0
}

func (c *ConcurrentDict) DeleteWithoutLock(key string) (val interface{}, result int) {
	shd := c.getShard(key)
	if val, ok := shd.m[key]; ok {
		delete(shd.m, key)
		c.subCount()
		return val, 1
	}
	return val, 0
}

func (c *ConcurrentDict) Put(key string, val interface{}) int {
	shd := c.getShard(key)
	shd.mu.Lock()
	defer shd.mu.Unlock()
	if _, ok := shd.m[key]; ok {
		shd.m[key] = val
		return 0
	}
	c.addCount()
	shd.m[key] = val
	return 1
}

func (c *ConcurrentDict) PutWithoutLock(key string, val interface{}) int {
	shd := c.getShard(key)
	if _, ok := shd.m[key]; ok {
		shd.m[key] = val
		return 0
	}
	c.addCount()
	shd.m[key] = val
	return 1
}

// PutIfAbsent 保存key( only insert)
func (c *ConcurrentDict) PutIfAbsent(key string, val interface{}) int {
	shd := c.getShard(key)
	shd.mu.Lock()
	defer shd.mu.Unlock()
	if _, ok := shd.m[key]; ok {
		return 0
	}
	c.addCount()
	shd.m[key] = val
	return 1
}

// PutIfAbsentWithLock 保存key( only insert)
func (c *ConcurrentDict) PutIfAbsentWithLock(key string, val interface{}) int {
	shd := c.getShard(key)
	if _, ok := shd.m[key]; ok {
		return 0
	}
	c.addCount()
	shd.m[key] = val
	return 1
}

// PutIfPresent 只更新
func (c *ConcurrentDict) PutIfPresent(key string, val interface{}) int {
	shd := c.getShard(key)
	shd.mu.Lock()
	defer shd.mu.Unlock()
	if _, ok := shd.m[key]; ok {
		shd.m[key] = val
		return 1
	}
	return 0
}

func (c *ConcurrentDict) PutIfPresentWithoutLock(key string, val interface{}) int {
	shd := c.getShard(key)
	if _, ok := shd.m[key]; ok {
		shd.m[key] = val
		return 1
	}
	return 0
}

func (c *ConcurrentDict) ForEach(consumer Consumer) {
	if c == nil {
		panic("dict cant be nil")
	}
	for _, sh := range c.shards {
		keep := sh.forEach(consumer)
		if !keep {
			break
		}
	}
}

// RWLock add only write lock
func (c *ConcurrentDict) RWLock(readKeys, writeKeys []string) {
	keys := append(readKeys, writeKeys...)
	allIndexes := c.toLockIndex(keys)
	writeIndexes := c.toLockIndexMap(writeKeys)
	for _, idx := range allIndexes {
		_, ok := writeIndexes[idx]
		rwMutex := &c.shards[idx].mu
		if !ok {
			rwMutex.RLock()
		} else {
			rwMutex.Lock()
		}
	}
}

func (c *ConcurrentDict) RWUnLock(readKeys, writeKeys []string) {
	keys := append(readKeys, writeKeys...)
	allIndexes := c.toLockIndex(keys)
	writeIndexes := c.toLockIndexMap(writeKeys)
	for _, idx := range allIndexes {
		_, ok := writeIndexes[idx]
		rwMutex := &c.shards[idx].mu
		if !ok {
			rwMutex.RUnlock()
		} else {
			rwMutex.Unlock()
		}
	}
}

func (c *ConcurrentDict) toLockIndex(keys []string) []uint32 {
	mapIndex := make(map[uint32]struct{})
	for _, key := range keys {
		mapIndex[c.index(utils.Fnv32(key))] = struct{}{}
	}
	indices := make([]uint32, 0, len(mapIndex))
	for k := range mapIndex {
		indices = append(indices, k)
	}
	sort.Slice(indices, func(i, j int) bool {
		return indices[i] < indices[j]
	})
	return indices
}

func (c *ConcurrentDict) toLockIndexMap(keys []string) map[uint32]struct{} {
	result := make(map[uint32]struct{})
	for _, key := range keys {
		result[c.index(utils.Fnv32(key))] = struct{}{}
	}
	return result
}
