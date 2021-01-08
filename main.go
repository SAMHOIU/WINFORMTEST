
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
					s := rand.Int()%20 + 5
					glog.Infof("queue is empty, worker %d sleep %d seconds", i, s)
					goutil.Sleep(time.Duration(s)*time.Second, exit)
				} else {
					glog.Error(err)
				}
				continue
			}
			task := new(et.UrlTask)
			if err = json.Unmarshal([]byte(item), task); err != nil {
				glog.Error(err)
				continue
			}

			var t time.Time
			bt, err := urlStore.Get(task.Url)
			if err == nil || err.Error() == "leveldb: not found" {
				t, _ = time.Parse(time.RFC3339, string(bt))
			}

			var req = &dl.HttpRequest{Url: task.Url}
			if *ua == "pc" || *ua == "mobile" || *ua == "google" {
				req.Platform = *ua
			}
			if *proxy {
				req.UseProxy = true
			}
			resp := dl.Download(req)
			if resp.Error != nil {
				glog.Error(resp.Error)
				continue
			}
			items := strings.Split(resp.RemoteAddr, ":")
			ip := ""
			if len(items) > 0 {
				ip = items[0]
			}
			tasks, records, err := Parse(task, resp.Text, ip)
			if err != nil {
				glog.Error(err)
				continue
			}

			t2 := time.Now()
			for _, rec := range records {
				b, _ := json.Marshal(rec)
				if *fs {
					fileStore.WriteLine(b)
				}
				if err = storeQueue.Enqueue(string(b)); err != nil {
					glog.Error(err)
				}

				if rec["time_"] != nil {
					switch rec["time_"].(type) {
					case string:
						t2, _ = time.Parse(time.RFC3339, rec["time_"].(string))
					}
				}
			}

			if t2.After(t) {
				for _, t := range tasks {
					if task.TaskName != "" {
						t.TaskName = task.TaskName
					}
					k := taskKey(t)
					if has, err := dedupStore.Has(k); has {
						continue
					} else if err != nil {
						glog.Error(err)
					}
					dedupStore.Put(k, nil)
					b, _ := json.Marshal(t)
					if err = crawlQueue.Enqueue(string(b)); err != nil {
						glog.Error(err)
					}
				}
			}
