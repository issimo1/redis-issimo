package interval

type Connection interface {
	GetDBIdx() int
	SetDBIdx(int)
	SetDBPwd(string)
	GetDBPwd() string
	Write([]byte) (int, error)

	IsClosed() bool
	Subscribe(channel string)
	UnSubscribe(channel string)
	SubscribeCount() int
	GetChannels() []string

	IsTransaction() bool
	SetTransaction(bool)

	EnqueueCmd([][]byte)
	GetQueueCmd() [][][]byte

	GetWatchKey() map[string]int64
	CleanWatchKey()

	AddTxErr(error)
	GetTxErr(error)
}
