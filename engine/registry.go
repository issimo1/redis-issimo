package engine

import (
	"github.com/issimo1/redis-issimo/redis/protocol"
	"strings"
)

type ExecFunc func(db *DB, args [][]byte) protocol.Reply
type KeysFunc func(args [][]byte) ([]string, []string)
type UndoFunc func(db *DB, args [][]byte) [][][]byte

var commandCenter map[string]*command = make(map[string]*command)

type command struct {
	commandName string
	execFunc    ExecFunc
	keyFunc     KeysFunc
	undoFunc    UndoFunc
	argsNum     int
}

func registerCommand(name string, args int, execFunc ExecFunc, keysFunc KeysFunc, undoFunc UndoFunc) {
	name = strings.ToLower(name)
	cmd := &command{}
	cmd.commandName = name
	cmd.argsNum = args
	cmd.execFunc = execFunc
	cmd.keyFunc = keysFunc
	cmd.undoFunc = undoFunc
	commandCenter[name] = cmd
}
