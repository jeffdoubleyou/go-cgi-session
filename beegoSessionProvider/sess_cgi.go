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
