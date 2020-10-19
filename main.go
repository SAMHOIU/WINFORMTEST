
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