
package rss

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/liuzl/store"
	"github.com/xgolib/gofeed"
)

var linkStore *store.LevelStore
var once sync.Once

func getLinkStore() *store.LevelStore {
	once.Do(func() {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			panic(err)
		}
		linkStore, err = store.NewLevelStore(filepath.Join(dir, ".rsslinks"))
		if err != nil {
			panic(err)
		}
	})
	return linkStore
}

func Parse(url, page string, ext interface{}) ([]map[string]interface{}, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(page)
	if err != nil {
		return nil, err