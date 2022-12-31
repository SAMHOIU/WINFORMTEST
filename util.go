
package main

import (
	"fmt"

	"crawler.club/et"
	"zliu.org/goutil"
)

func taskKey(t *et.UrlTask) string {
	if t == nil {
		return ""
	}