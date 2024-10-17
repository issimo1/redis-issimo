package config

import (
	"bufio"
	"flag"
	"github.com/issimo1/redis-issimo/utils/logger"
	"github.com/issimo1/redis-issimo/utils/rand"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	runidMaxLen          = 40
	defaultDatabaseCount = 16
)

type RedisConfig struct {
	Bind           string   `conf:"bind"`
	Port           int      `conf:"port"`
	Dir            string   `conf:"dir"`
	RunId          string   `conf:"runId"`
	DBCount        int      `conf:"dbCounts"`
	AppendOnly     bool     `conf:"appendOnly"`
	AppendFileName string   `conf:"appendFileName"`
	AppendFsync    string   `conf:"appendFsync"`
	RequiredPwd    string   `conf:"requiredPwd,omitempty"`
	Cluster        []string `conf:"cluster"`
	Self           string   `conf:"self"`
}

var (
	ConfigPath   = ""
	GlobalConfig *RedisConfig
)

func defaultRedisConfig() {
	GlobalConfig = &RedisConfig{
		Bind:       "127.0.0.1",
		Port:       6379,
		Dir:        ".",
		RunId:      rand.RandString(runidMaxLen),
		DBCount:    defaultDatabaseCount,
		AppendOnly: false,
	}
}

func LoadConfig(config string) {
	conf, err := os.Open(config)
	if err != nil {
		logger.Warnf("the file %s is not exists. using defaultRedisConfig().", config)
		defaultRedisConfig()
		return
	}
	defer conf.Close()
	GlobalConfig = parse(conf)

}

func parse(r io.Reader) *RedisConfig {
	newRedisConf := &RedisConfig{}

	lineMap := make(map[string]string)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimLeft(line, " ")

		if len(line) == 0 || (len(line) > 0 && line[0] == '#') {
			continue
		}

		idx := strings.IndexAny(line, " ")
		if idx > 0 && idx < len(line)-1 {
			key := line[:idx][:len(line[:idx])-1]
			value := strings.Trim(line[idx+1:], " ")
			lineMap[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err.Error())
	}

	confValue := reflect.ValueOf(newRedisConf).Elem()
	confType := reflect.TypeOf(newRedisConf).Elem()
	for i := 0; i < confType.NumField(); i++ {
		field := confType.Field(i)
		fieldName := strings.Trim(field.Tag.Get("conf"), " ")
		if fieldName == "" {
			fieldName = field.Name
		} else {
			fieldName = strings.Split(fieldName, ",")[0]
		}
		//fieldName = strings.ToLower(fieldName)
		fieldValue, ok := lineMap[fieldName]
		if ok {
			switch field.Type.Kind() {
			case reflect.String:
				confValue.Field(i).SetString(fieldValue)
			case reflect.Bool:
				confValue.Field(i).SetBool(fieldValue == "yes")
			case reflect.Int:
				intValue, err := strconv.ParseInt(fieldValue, 10, 64)
				if err == nil {
					confValue.Field(i).SetInt(intValue)
				}
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.String {
					tmpSlice := strings.Split(fieldValue, ",")
					confValue.Field(i).Set(reflect.ValueOf(tmpSlice))
				}
			}
		}
	}
	return newRedisConf
}

func Init() {
	flag.StringVar(&ConfigPath, "conf", "./../../redis.conf", "config file path. default conf=./../redis.conf")
	flag.Parse()
	LoadConfig(ConfigPath)
}
