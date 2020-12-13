
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
	once.Do(func() {
		crawlDir := filepath.Join(*dir, "crawl")
		if crawlQueue, err = q.NewQueueWithRetryLimit(crawlDir, *retry); err != nil {
			glog.Error(err)
			return
		}
		storeDir := filepath.Join(*dir, "store")
		if storeQueue, err = q.NewQueue(storeDir); err != nil {
			glog.Error(err)
			return
		}
		dbDir := filepath.Join(*dir, "url")
		if urlStore, err = store.NewLevelStore(dbDir); err != nil {
			glog.Error(err)
			return
		}
		dedupDir := filepath.Join(*dir, "dedup")
		if dedupStore, err = store.NewLevelStore(dedupDir); err != nil {
			glog.Error(err)
			return
		}
		if *fs {
			fsDir := filepath.Join(*dir, "fs")
			if fileStore, err = filestore.NewFileStore(fsDir); err != nil {
				glog.Error(err)
				return
			}
		}
		if goutil.FileGuard("first.lock") {
			if err = initSeeds(); err != nil {
				return
			}
		}
	})
	return
}

func initSeeds() error {
	seedsFile := filepath.Join(*conf, "seeds.json")
	content, err := ioutil.ReadFile(seedsFile)
	if err != nil {
		glog.Error(err)
		return err
	}
	var seeds []*et.UrlTask
	if err = json.Unmarshal(content, &seeds); err != nil {
		glog.Error(err)
		return err
	}
	glog.Infof("initSeeds %d seeds", len(seeds))
	tz := time.Now().Format("200601020304")
	for _, seed := range seeds {
		seed.TaskName = tz
		b, _ := json.Marshal(seed)
		if err = crawlQueue.Enqueue(string(b)); err != nil {
			glog.Error(err)
			return err
		}
	}
	return nil
}

func stop(sigs chan os.Signal, exit chan bool) {
	<-sigs
	glog.Info("receive stop signal!")
	close(exit)
}

func work(i int, exit chan bool) {
	glog.Infof("start worker %d", i)
	for {
		select {
		case <-exit:
			glog.Infof("worker %d exit", i)
			return
		default:
			key, item, err := crawlQueue.Dequeue(*timeout)
			if err != nil {
				if err.Error() == "Queue is empty" {