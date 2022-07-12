package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultRedisAddr = "localhost:6379"
	defaultPort      = "8388"
	defaultHost      = "localhost"
)

// Options is the option to run the application
type Options struct {
	Host   string        `json:"host" env:"APP_HOST"`
	Domain string        `json:"domain" env:"APP_DOMAIN"`
	Port   string        `json:"port" env:"APP_PORT"`
	Expiry time.Duration `json:"duration" env:"APP_EXPIRY"`
	Redis  RedisOption   `json:"redis,omitempty"`
	Badger BadgerOption  `json:"badger,omitempty"`
}

func New(getOptionFn func() Options) Options {
	o := getOptionFn()
	if o.Redis.Addresses == nil {
		o.Redis.Addresses = []string{defaultRedisAddr}
	}

	if o.Host == "" {
		o.Host = defaultHost
	}

	if o.Port == "" {
		o.Port = defaultPort
	}

	if o.Domain == "" {
		o.Domain = fmt.Sprintf("http://" + o.Host + ":" + o.Port)
	}

	if o.Badger.Path == "" {
		o.Badger.Path = filepath.Join(os.TempDir(), "shortener-badger")
	}

	return o
}

// RedisOption contains addresses to be used for redis
type RedisOption struct {
	Addresses []string `json:"addresses"`
}

func (r RedisOption) Validate() error {
	return nil
}

// BadgerOption is the option for badger
type BadgerOption struct {
	Path string `json:"path"`
}

func (b BadgerOption) Validate() error {
	if _, err := os.Stat(b.Path); err != nil {
		return err
	}
	return nil
}
