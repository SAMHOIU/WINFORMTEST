
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"crawler.club/et"
	"github.com/golang/glog"
	"zliu.org/goutil/rest"
)

var (
	addr = flag.String("addr", ":2001", "rest address")
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infof("addr=%s  method=%s host=%s uri=%s",