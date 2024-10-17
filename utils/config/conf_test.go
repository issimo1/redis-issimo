package config

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	Init()
	fmt.Println(GlobalConfig)
}
