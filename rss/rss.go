
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