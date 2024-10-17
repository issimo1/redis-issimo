package rand

import (
	"math/rand"
	"time"
)

// 随机对象*rand
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandString(n int) string {
	result := make([]rune, n)
	for i := 0; i < len(result); i++ {
		result[i] = rune(letters[r.Intn(len(letters))])
	}
	return string(result)
}
