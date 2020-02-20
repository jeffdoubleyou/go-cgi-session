package drivers

import (
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedDriver struct {
	client *memcache.Client
}

type MemcachedDriverConfig struct {
	servers []string
	timeout int
}

func Memcached(c interface{}) *MemcachedDriver {
	driverConfig := &MemcachedDriverConfig{}
	config := c.(map[string]interface{})

	if servers, ok := config["servers"]; ok {
		driverConfig.servers = servers.([]string)
	} else {
		driverConfig.servers = []string{"127.0.0.1:11211"}
	}

	if timeout, ok := config["timeout"]; ok {
		driverConfig.timeout = timeout.(int)
	} else {
		driverConfig.timeout = 0
	}

	client := memcache.New(driverConfig.servers...)
	if driverConfig.timeout != 0 {
		client.Timeout, _ = time.ParseDuration(fmt.Sprintf("%ds", driverConfig.timeout))
	}
	return &MemcachedDriver{client}
}

func (d *MemcachedDriver) Store(session string, params []byte) (store []byte, err error) {
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
