package drivers

import(
    "fmt"
    "time"
    "github.com/bradfitz/gomemcache/memcache"
)

type MemcachedDriver struct {
    client *memcache.Client
}

func Memcached(config map[string]interface{}) *MemcachedDriver {
    if config["servers"] == nil {
        config["servers"] = []string{"127.0.0.1:11211"}
    }
    client := memcache.New(config["servers"].([]string)...)
    if config["timeout"] != nil {
        client.Timeout, _ = time.ParseDuration(fmt.Sprintf("%ds", config["timeout"].(int)))
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


