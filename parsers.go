
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

func (p *Parsers) GetParser(name string, refresh bool) (*et.Parser, error) {
	p.Lock()
	defer p.Unlock()
	if !refresh && p.items[name] != nil {
		return p.items[name], nil
	}
	file := filepath.Join(*conf, "parsers", name+".json")
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	parser := new(et.Parser)
	if err := json.Unmarshal(content, parser); err != nil {
		return nil, err
	}
	p.items[name] = parser
	return parser, nil
}

var pool = &Parsers{items: make(map[string]*et.Parser)}

func Parse(task *et.UrlTask, page, ip string) (
	[]*et.UrlTask, []map[string]interface{}, error) {
	name := task.ParserName