package net_log

import (
	"fmt"
	"time"
	"runtime/debug"
)

func Logout(cate string, content string, others ... string) {
	str := time.Now().Format("01-01 01:01:01")
	content1 := fmt.Sprintf(content, others)
	logStr := fmt.Sprintf("[%s][%s] %s", cate, str, content1)
	fmt.Println(logStr)
	debug.PrintStack()
}
