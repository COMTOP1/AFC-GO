package main

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type (
	Config struct {
		Address    string
		DomainName string
	}

	// Views encapsulates our view dependencies
	Views struct {
		cache    *cache.Cache
		conf     *Config
		template *Templater
	}
)

func New(conf *Config) *Views {
	v := &Views{}

	v.template = NewTemplate()

	// Initialising cache
	v.cache = cache.New(1*time.Hour, 1*time.Hour)

	v.conf = conf

	return v
}
