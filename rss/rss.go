
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
