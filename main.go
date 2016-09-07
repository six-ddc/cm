package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"regexp"
	"time"
)

func Errorf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf("[\x1b[0;31mERROR\x1b[0;0m] %s\n", fmt.Sprintf(format, a...))
}

func SErrorf(format string, a ...interface{}) string {
	return fmt.Sprintf("\x1b[0;31m%s\x1b[0;0m", fmt.Sprintf(format, a...))
}

func Errorln(a ...interface{}) (n int, err error) {
	return fmt.Printf("[\x1b[0;31mERROR\x1b[0;0m] %s", fmt.Sprintln(a...))
}

func SErrorln(a ...interface{}) string {
	return fmt.Sprintf("\x1b[0;31m%s\x1b[0;0m", fmt.Sprint(a...))
}

func Infof(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf("[\x1b[0;32mINFOO\x1b[0;0m] %s\n", fmt.Sprintf(format, a...))
}

func SInfof(format string, a ...interface{}) string {
	return fmt.Sprintf("\x1b[0;32m%s\x1b[0;0m", fmt.Sprintf(format, a...))
}

func Infoln(a ...interface{}) (n int, err error) {
	return fmt.Printf("[\x1b[0;32mINFOO\x1b[0;0m] %s", fmt.Sprintln(a...))
}

func SInfoln(a ...interface{}) string {
	return fmt.Sprintf("\x1b[0;32m%s\x1b[0;0m", fmt.Sprint(a...))
}

func Debugf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf("[\x1b[0;34mDEBUG\x1b[0;0m] %s\n", fmt.Sprintf(format, a...))
}

func SDebugf(format string, a ...interface{}) string {
	return fmt.Sprintf("\x1b[0;34m%s\x1b[0;0m", fmt.Sprintf(format, a...))
}

func Debugln(a ...interface{}) (n int, err error) {
	return fmt.Printf("[\x1b[0;34mDEBUG\x1b[0;0m] %s", fmt.Sprintln(a...))
}

func SDebugln(a ...interface{}) string {
	return fmt.Sprintf("\x1b[0;34m%s\x1b[0;0m", fmt.Sprint(a...))
}

type RediConf struct {
	Address string
	Conn    redis.Conn
}

func NewRediConf(address string) (*RediConf, error) {
	cf := &RediConf{
		Address: address,
	}
	var err error
	cf.Conn, err = redis.Dial("tcp", cf.Address)
	return cf, err
}

func (cf *RediConf) AddValue(key, val string) (bool, error) {
	ts := time.Now().Unix()
	ok := true
	_, err := redis.String(cf.Conn.Do("hmset", key, "value", val, "ts", ts))
	if err != nil {
		Errorf("HMSET %s failed. %s", key, err)
		ok = false
	}
	return ok, err
}

func (cf *RediConf) GetValue(key string) (string, error) {
	ret, err := redis.StringMap(cf.Conn.Do("hgetall", key))
	if err != nil {
		return "", err
	}
	val, ok := ret["value"]
	if !ok {
		return "", errors.New("Not exists value field")
	}
	return val, nil
}

func (cf *RediConf) Close() {
	cf.Conn.Close()
}

var (
	redisConf   = kingpin.Flag("redis", "Redis server address.").Default("127.0.0.1:6379").String()
	redisPrefix = kingpin.Flag("redis-prefix", "Prefix of redis key.").Default("").String()

	getMode    = kingpin.Command("get", "Get mode.")
	outputFile = getMode.Flag("output", "Output file name.").Default("<input-file>.out").String()
	inputFile  = getMode.Arg("input-file", "Input template config file.").Required().ExistingFile()

	setMode = kingpin.Command("set", "Set mode.")
	setPair = setMode.Arg("key=value", "Save key value pair.").Required().StringMap()

	/*
		serverMode = kingpin.Command("server", "Server mode.")
		serverPort = serverMode.Flag("port", "Server listen port.").Default("10086").Uint16()
	*/
)

func doGetCommand() int {

	cf, err := NewRediConf(*redisConf)
	if err != nil {
		Errorln(err)
		return 1
	}
	defer cf.Close()
	Debugf("Connect redis:%s succesed", *redisConf)

	fp, err := os.Open(*inputFile)
	if err != nil {
		Errorf("Open %s failed. %s", *inputFile, err)
		return 1
	}
	defer fp.Close()

	if *outputFile == "<input-file>.out" {
		*outputFile = *inputFile + ".out"
	}
	wfp, err := os.OpenFile(*outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		Errorf("Open %s failed. %s", *outputFile, err)
		return 1
	}
	defer wfp.Close()

	buff := bufio.NewReader(fp)
	wbuff := bufio.NewWriter(wfp)
	re, err := regexp.Compile("\\$\\{[_a-zA-Z0-9\\-\\.:]+\\}")

	var failKeys, succKeys []string
	for true {
		line, _, err := buff.ReadLine()
		if err != nil {
			break
		}
		newLine := re.ReplaceAllFunc(line, func(str []byte) []byte {
			k := string(str[2 : len(str)-1])
			if len(*redisPrefix) > 0 {
				k = *redisPrefix + ":" + k
			}
			val, err := cf.GetValue(k)
			if err == nil {
				succKeys = append(succKeys, k)
				return []byte(val)
			} else {
				failKeys = append(failKeys, k)
				return str
			}
		})
		wbuff.Write(newLine)
		wbuff.WriteString("\n")
	}
	wbuff.Flush()

	if len(succKeys) > 0 {
		Infoln("get succ:", SInfoln(succKeys))
	}

	if len(failKeys) > 0 {
		Errorln("get fail:", SErrorln(failKeys))
		return 1
	} else {
		return 0
	}
}

func doSetCommand() int {

	cf, err := NewRediConf(*redisConf)
	if err != nil {
		Errorln(err)
		return 1
	}
	defer cf.Close()

	Debugf("Connect redis:%s succesed", *redisConf)

	var failKeys, succKeys []string
	for k, v := range *setPair {
		if len(*redisPrefix) > 0 {
			k = *redisPrefix + ":" + k
		}
		ok, _ := cf.AddValue(k, v)
		if ok {
			succKeys = append(succKeys, k)
		} else {
			failKeys = append(failKeys, k)
		}
	}

	if len(succKeys) > 0 {
		Infoln("set succ:", SInfoln(succKeys))
	}

	if len(failKeys) > 0 {
		Errorln("set fail:", SErrorln(failKeys))
		return 1
	} else {
		return 0
	}
}

func main() {

	kingpin.Version("0.0.1")
	kingpin.CommandLine.Name = "qconf"
	kingpin.CommandLine.Help = "A command-line configuration management tool"

	if len(os.Args) == 1 {
		kingpin.Usage()
		return
	}

	ret := 1

	switch kingpin.Parse() {
	case getMode.FullCommand():
		ret = doGetCommand()
	case setMode.FullCommand():
		ret = doSetCommand()
	}

	os.Exit(ret)
}
