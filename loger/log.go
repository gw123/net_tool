package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	// 使用 JSONFormatter
	log.SetFormatter(&log.JSONFormatter{})
	// 使用 logrus 像往常那样记录事件
	log.WithFields(log.Fields{"string": "foo", "int": 1, "float": 1.1}).Info("My first ssl event from golang")
}
