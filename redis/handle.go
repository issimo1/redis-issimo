package redis

import (
	"context"
	"errors"
	"github.com/issimo1/redis-issimo/abs"
	"github.com/issimo1/redis-issimo/redis/connection"
	"github.com/issimo1/redis-issimo/redis/parser"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/issimo1/redis-issimo/utils/config"
	"github.com/issimo1/redis-issimo/utils/logger"
)

type Handler struct {
	activeConn *sync.Map
	engine     abs.Engine
}

func NewDefaultHandler() *Handler {
	var e abs.Engine
	if len(config.GlobalConfig.Cluster) > 0 {
		logger.Info("cluster mode")

	} else {
		logger.Info("single mode")
	}

	return &Handler{
		engine: e,
	}
}

func (h *Handler) Handle(c context.Context, conn net.Conn) {
	keepConn := connection.NewKeepConnection(conn)
	h.activeConn.Store(conn, struct{}{})
	outChan := parser.ParseStream(conn)
	for payload := range outChan {
		if payload.Err != nil {
			if payload.Err == io.EOF || errors.Is(payload.Err, io.ErrUnexpectedEOF) ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				h.activeConn.Delete(conn)
				logger.Warn("client closed:" + keepConn.C.RemoteAddr().String())
				keepConn.C.Close()
				return
			}
		}
		logger.Debug("%q")
		result := h.engine.Exec(keepConn, nil)
		if result != nil {
			keepConn.Write(result.ToBytes())
		} else {
			keepConn.Write(nil)
		}
	}
}
