package beegoSessionProvider

import (
	"encoding/json"
	"fmt"
	"log"
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
	config      map[string]interface{}
	session     *cgisession.CGISession
	test        string
}

func (p *CGIProvider) SessionInit(maxlifetime int64, config string) error {
	p.maxlifetime = maxlifetime
	var c interface{}
	err := json.Unmarshal([]byte(config), &c)
	if err != nil {
		return err
	}
	log.Printf("THE CONFIG IS HERE: %v\n\n", c)
	p.config = c.(map[string]interface{})
	p.test = "THIS IS A STRING"
	return nil
}

func (p *CGIProvider) SessionRead(sid string) (session.Store, error) {
	log.Printf("Reading session now: %s\n\n", p.test)
	log.Printf("#### THIS IS THE CONFIG ####\n%v\n########\n", p.config)
	p.session = cgisession.Session(p.config)
	if params := p.session.Load(sid); params == nil {
		return nil, fmt.Errorf("Invalid session")
	}
	session := &SessionStore{sid: sid, session: p.session}
	return session, nil
}

func (p *CGIProvider) SessionExist(sid string) bool {
	if p.session == nil {
		p.session = cgisession.Session(p.config)
	}
	return p.session.Exists(sid)
}

func (p *CGIProvider) SessionRegenerate(oldsid, sid string) (session.Store, error) {
	if p.session == nil {
		p.session = cgisession.Session(p.config)
	}
	if params := p.session.New(oldsid, sid); params == nil {
		return nil, fmt.Errorf("Invalid session")
	}
	session := &SessionStore{sid: sid, session: p.session}
	return session, nil
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
