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

package beegoSessionProvider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	session "github.com/astaxie/beego/session"
	cgisession "github.com/jeffdoubleyou/go-cgi-session"
)

var cgipder = &CGIProvider{}

type SessionStore struct {
	sid     string
	lock    sync.RWMutex
	session *cgisession.SessionStore
}

func (s *SessionStore) Set(key, value interface{}) error {
	s.session.SetParam(key, value)
	return nil
}

func (s *SessionStore) Get(key interface{}) interface{} {
	return s.session.GetParam(key)
}

func (s *SessionStore) Delete(key interface{}) error {
	s.session.ClearParam(key.(string))
	return nil
}

func (s *SessionStore) Flush() error {
	return nil
}

func (s *SessionStore) SessionID() string {
	return s.session.SessionId()
}

func (s *SessionStore) SessionRelease(w http.ResponseWriter) {
	s.session.Flush()
}

type CGIProvider struct {
	maxlifetime int64
	config      *cgisession.CGISessionConfig
	session     *cgisession.CGISession
}

func (p *CGIProvider) SessionInit(maxlifetime int64, config string) error {
	p.maxlifetime = maxlifetime
	c := &cgisession.CGISessionConfig{}
	err := json.Unmarshal([]byte(config), &c)
	if err != nil {
		return err
	}
	p.config = c
	p.session = cgisession.Session(p.config)
	return nil
}

func (p *CGIProvider) SessionRead(sid string) (session.Store, error) {
	if s := p.session.New(sid); s == nil {
		return nil, fmt.Errorf("Invalid session %s", sid)
	} else {
		return &SessionStore{sid: s.SessionId(), session: s}, nil
	}
}

func (p *CGIProvider) SessionExist(sid string) bool {
	if p.session == nil {
		p.session = cgisession.Session(p.config)
	}
	return p.session.Exists(sid)
}

func (p *CGIProvider) SessionRegenerate(oldsid, sid string) (session.Store, error) {
	if s := p.session.New(oldsid, sid); s == nil {
		return nil, fmt.Errorf("Invalid session")
	} else {
		return &SessionStore{sid: sid, session: s}, nil
	}
}

func (p *CGIProvider) SessionDestroy(sid string) error {
	return nil
}

func (p *CGIProvider) SessionGC() {

}

func (p *CGIProvider) SessionAll() int {
	return 1
}

func init() {
	session.Register("cgi", cgipder)
}
