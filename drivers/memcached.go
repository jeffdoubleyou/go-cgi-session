// Copyright 2020 Jeffrey Weitz &lt;jeffdoubleyou@gmail.com&gt;
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drivers

import (
	"fmt"
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
	client := memcache.New(driverConfig.Servers...)
	if driverConfig.Timeout != 0 {
		client.Timeout, _ = time.ParseDuration(fmt.Sprintf("%ds", driverConfig.Timeout))
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
