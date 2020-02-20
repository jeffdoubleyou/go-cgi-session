package drivers

import (
	"fmt"
	"log"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedDriver struct {
	client *memcache.Client
}

type DriverConfig struct {
	Servers []string
	Timeout int
}

func Memcached(c ...*DriverConfig) *MemcachedDriver {
	var driverConfig *DriverConfig

	if len(c) == 1 && c[0] != nil {
		driverConfig = c[0]
	} else {
		driverConfig = &DriverConfig{Servers: []string{"127.0.0.1:11211"}, Timeout: 0}
	}
	log.Printf("SERVER LIST: %v\n", driverConfig.Servers)
	client := memcache.New(driverConfig.Servers...)
	if driverConfig.Timeout != 0 {
		client.Timeout, _ = time.ParseDuration(fmt.Sprintf("%ds", driverConfig.Timeout))
	}
	return &MemcachedDriver{client}
}

func (d *MemcachedDriver) Store(session string, params []byte) (store []byte, err error) {
	log.Printf("Request to store session ID: %s\n******\n%v\n\n", session, params)
	d.client.Set(&memcache.Item{Key: session, Value: params})
	return store, err
}

func (d *MemcachedDriver) Retrieve(session string) (data []byte, err error) {
	item, err := d.client.Get(session)
	if err == nil {
		data = item.Value
	}
	return data, err
}

func (d *MemcachedDriver) Remove(session string) (status bool, err error) {
	status = true
	d.client.Delete(session)
	return status, err
}
