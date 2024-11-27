package connection

import (
	"github.com/issimo1/redis-issimo/utils/logger"
	"net"
	"sync"
	"sync/atomic"
)

var connPoll = sync.Pool{
	New: func() interface{} {
		return &KeepConnection{
			dbIdx:    0,
			C:        nil,
			password: "",
			closed:   atomic.Bool{},
			tx:       atomic.Bool{},
		}
	},
}

type KeepConnection struct {
	C                  net.Conn
	password           string
	dbIdx              int
	WriteDateWaitGroup sync.WaitGroup

	mu     sync.Mutex
	subs   map[string]struct{}
	closed atomic.Bool

	tx        atomic.Bool
	queue     [][][]byte
	watchKey  map[string]int64
	txErrList []error
}

func NewKeepConnection(c net.Conn) *KeepConnection {
	conn, ok := connPoll.Get().(*KeepConnection)
	if !ok {
		logger.Error("connection pool make wrong type")
		return &KeepConnection{
			dbIdx:    0,
			C:        nil,
			password: "",
			closed:   atomic.Bool{},
			tx:       atomic.Bool{},
		}
	}
	conn.C = c
	conn.closed.Store(false)
	conn.tx.Store(false)
	conn.queue = nil
	conn.txErrList = nil
	conn.watchKey = nil
	return conn
}

func (k *KeepConnection) GetDBPwd() string {
	return k.password
}

func (k *KeepConnection) Write(bytes []byte) (int, error) {
	if len(bytes) == 0 {
		return 0, nil
	}
	k.WriteDateWaitGroup.Add(1)
	defer k.WriteDateWaitGroup.Done()
	return k.C.Write(bytes)
}

func (k *KeepConnection) IsClosed() bool {
	return k.closed.Load()
}

func (k *KeepConnection) Subscribe(channel string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.subs == nil {
		k.subs = make(map[string]struct{})
	}
	k.subs[channel] = struct{}{}
}

func (k *KeepConnection) UnSubscribe(channel string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if len(k.subs) == 0 {
		return
	}
	delete(k.subs, channel)
}

func (k *KeepConnection) SubscribeCount() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.subs)
}

func (k *KeepConnection) GetChannels() []string {
	k.mu.Lock()
	defer k.mu.Unlock()
	channels := make([]string, 0)
	for idx := range k.subs {
		channels = append(channels, idx)
	}
	return channels
}

func (k *KeepConnection) IsTransaction() bool {
	return k.tx.Load()
}

func (k *KeepConnection) SetTransaction(b bool) {
	if !b {
		k.queue = nil
		k.watchKey = nil
		k.txErrList = nil
	}
	k.tx.Store(b)
}

func (k *KeepConnection) EnqueueCmd(i [][]byte) {
	k.queue = append(k.queue, i)
}

func (k *KeepConnection) GetQueueCmd() [][][]byte {
	return k.queue
}

func (k *KeepConnection) GetWatchKey() map[string]int64 {
	if k.watchKey == nil {
		k.watchKey = make(map[string]int64)
	}
	return k.watchKey
}

func (k *KeepConnection) CleanWatchKey() {
	k.watchKey = nil
}

func (k *KeepConnection) AddTxErr(err error) {
	//TODO implement me
	panic("implement me")
}

func (k *KeepConnection) GetTxErr(err error) {
	//TODO implement me
	panic("implement me")
}

func (k *KeepConnection) GetDBIdx() int {
	return k.dbIdx
}

func (k *KeepConnection) SetDBIdx(idx int) {
	k.dbIdx = idx
}

func (k *KeepConnection) SetDBPwd(pwd string) {
	k.password = pwd
}
