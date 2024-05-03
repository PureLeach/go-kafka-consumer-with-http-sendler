package utils

import (
	"github.com/jellydator/ttlcache/v3"
)

var CacheMain = ttlcache.New[string, string]()
