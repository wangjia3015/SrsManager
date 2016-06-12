package utils

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"os"
	"time"
)

var workerNameStr string

func init() {
	hn, err := os.Hostname()
	if err != nil {
		hn = os.Getenv("HOSTNAME")
	}
	if len(hn) == 0 {
		hn = "localhost"
	}

	workerNameStr = fmt.Sprintf("%s-%d", hn, os.Getpid())
}

func GenerateUuid() string {
	hasher := md5.New()
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if err != nil || n != len(uuid) {
		src := mrand.NewSource(time.Now().UnixNano())
		r := mrand.New(src)
		for n := range uuid {
			uuid[n] = byte(r.Int())
		}
	}

	hasher.Write([]byte(workerNameStr))
	hasher.Write(uuid)
	hasher.Write([]byte(time.Now().String()))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
