
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"crawler.club/ce"
	"crawler.club/crawler/rss"
	"crawler.club/et"
)

var (
	conf = flag.String("conf", "./conf", "dir for parsers conf")
)

type Parsers struct {
	sync.Mutex
	items map[string]*et.Parser
}