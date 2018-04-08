package main

import (
	"crypto/md5"
	"fmt"
)

func Hash(data string, salt string) string {
	output := md5.Sum([]byte(data + salt))
	return fmt.Sprintf("%x", output)
}
