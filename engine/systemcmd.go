package engine

import (
	"github.com/issimo1/redis-issimo/abs"
	"github.com/issimo1/redis-issimo/redis/protocol"
	"github.com/issimo1/redis-issimo/utils/config"
)

func Auth(c abs.Connection, pwd [][]byte) protocol.Reply {
	if len(pwd) != 1 {
		//err
	}
	if config.GlobalConfig.RequiredPwd == "" {
		// err
	}
	pwds := string(pwd[0])
	if config.GlobalConfig.RequiredPwd != pwds {
		// err
	}
	c.SetDBPwd(pwds)
	return nil
}

// Ping todo protocol
func Ping(ping [][]byte) protocol.Reply {
	if len(ping) == 0 {
		return nil
	} else if len(ping) == 1 {
		//return ping[0]
	}
	// err
	return nil
}

func checkPassword(c abs.Connection) bool {
	if config.GlobalConfig.RequiredPwd == "" {
		return true
	}
	return c.GetDBPwd() == config.GlobalConfig.RequiredPwd
}
