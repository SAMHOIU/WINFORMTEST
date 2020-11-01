
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"crawler.club/crawler/version"
	"crawler.club/dl"
	"crawler.club/et"
	"github.com/golang/glog"
	"github.com/liuzl/store"
	"zliu.org/filestore"
	"zliu.org/goutil"
	"zliu.org/q"
)

var (
	dir     = flag.String("dir", "data", "working dir")
	timeout = flag.Int64("timeout", 300, "in seconds")
	c       = flag.Int("c", 1, "worker count")
	retry   = flag.Int("retry", 5, "retry cnt")
	period  = flag.Int("period", -1, "period in seconds")
	fs      = flag.Bool("fs", true, "filestore flag")
	api     = flag.Bool("api", false, "http api flag")
	proxy   = flag.Bool("proxy", false, "use proxy or not")
	ua      = flag.String("ua", "", "pc, mobile, google. Golang UA for empty")
)

var crawlQueue, storeQueue *q.Queue
var urlStore, dedupStore *store.LevelStore
var fileStore *filestore.FileStore
var once sync.Once

func finish() {
	if crawlQueue != nil {
		crawlQueue.Close()
	}
	if storeQueue != nil {
		storeQueue.Close()
	}
	if urlStore != nil {
		urlStore.Close()
	}
	if dedupStore != nil {
		dedupStore.Close()
	}
	if fileStore != nil {
		fileStore.Close()
	}
}

func initTopics() (err error) {