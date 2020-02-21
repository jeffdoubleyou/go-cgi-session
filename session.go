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

package cgisession

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jeffdoubleyou/go-cgi-session/drivers"
	sessionid "github.com/jeffdoubleyou/go-cgi-session/id"
	"github.com/jeffdoubleyou/go-cgi-session/serializers"
)

const (
	VERSION = "0.1.1"
)

type CGISessionConfig struct {
	Driver           string
	DriverConfig     *drivers.DriverConfig
	Id               string
	IdConfig         interface{}
	Serializer       string
	SerializerConfig interface{}
	ExpireSeconds    int64
	IPMatch          bool
}

type driver interface {
	Retrieve(session string) ([]byte, error)
	Store(session string, params []byte) ([]byte, error)
	Remove(session string) (bool, error)
}

type id interface {
	Generate() string
}

type serializer interface {
	Freeze(params interface{}) ([]byte, error)
	Thaw(params []byte) (interface{}, error)
}

type CGISession struct {
	config     *CGISessionConfig
	driver     driver
	id         id
	serializer serializer
}

type SessionStore struct {
	session   *CGISession
	sessionId string
	params    map[string]interface{}
}

func Session(config ...*CGISessionConfig) *CGISession {
	var sessionConfig *CGISessionConfig
	if len(config) == 1 {
		sessionConfig = config[0]
	} else {
		sessionConfig = &CGISessionConfig{Driver: "memcached", Id: "md5", Serializer: "datadumper", ExpireSeconds: 86400, IPMatch: false}
	}
	return &CGISession{config: sessionConfig, driver: nil, id: nil, serializer: nil}
}

func (s *CGISession) Id(idgen ...id) {
	if s.id == nil {
		if len(idgen) == 1 {
			s.id = idgen[0]
		} else {
			switch idGen := s.config.Id; idGen {
			case "md5":
				s.id = sessionid.Md5()
			default:
				s.id = sessionid.Md5()
			}
		}
	}
}

func (s *CGISession) GenerateSessionId() string {
	s.Id()
	return s.id.Generate()
}

func (s *CGISession) Serializer(sessionSerializer ...serializer) {
	if s.serializer == nil {
		if len(sessionSerializer) == 1 {
			s.serializer = sessionSerializer[0]
		} else {
			switch ss := s.config.Serializer; ss {
			case "datadumper":
				s.serializer = serializers.DataDumper(s.config.SerializerConfig)
			default:
				s.serializer = serializers.DataDumper(s.config.SerializerConfig)
			}
		}
	}
}

func (s *CGISession) Thaw(data []byte) (map[string]interface{}, error) {
	s.Serializer()
	params, err := s.serializer.Thaw(data)
	if err != nil {
		return nil, err
	}
	return params.(map[string]interface{}), nil
}

func (s *CGISession) Freeze(params interface{}) (data []byte, err error) {
	s.Serializer()
	data, err = s.serializer.Freeze(params)
	return
}

func (s *CGISession) Driver(sessionDriver ...driver) {
	if s.driver == nil {
		if len(sessionDriver) == 1 {
			s.driver = sessionDriver[0]
		} else {
			switch d := s.config.Driver; d {
			case "memcached":
				s.driver = drivers.Memcached(s.config.DriverConfig)
			default:
				s.driver = drivers.Memcached(s.config.DriverConfig)
			}
		}
	}
}

// Beego seems to insist on choosing it's own session ID, so we must accept a session ID here, although I don't fine this secure, since you could change someone's login to be pointed at your session information if you knew their session ID
func (s *CGISession) New(sessionId ...string) (ss *SessionStore) {
	s.Driver()
	if len(sessionId) > 0 {
		var newSessionId string
		if len(sessionId) == 2 {
			newSessionId = sessionId[1]
		} else {
			newSessionId = sessionId[0]
		}
		ss = s.Load(sessionId[0])
		if ss == nil {
			ss = s.createSession(newSessionId)
		} else {
			// Replace session with new session ID
			if newSessionId != sessionId[0] {
				s.deleteSession(sessionId[0])
				ss.sessionId = newSessionId
				now := time.Now().Unix()
				ss.ParamInt64("_SESSION_CTIME", now)
				ss.ParamInt64("_SESSION_ATIME", now)
				ss.ParamInt64("_SESSION_ETIME", s.config.ExpireSeconds)
				ss.Flush()
			}
		}
	} else {
		ss = s.createSession()
	}
	return ss
}

func (s *CGISession) Load(sessionId string) (ss *SessionStore) {
	s.Driver()
	ss = &SessionStore{session: s}
	data, err := s.driver.Retrieve(sessionId)
	if err != nil {
		return nil
	} else {
		ss.params, err = s.Thaw(data)
		if err != nil {
			return nil
		}
		if ss.IPMatches() == false {
			log.Printf("Deleteing session not matching the IP address")
			s.deleteSession(sessionId)
			return nil
		}
		if ss.IsExpired() == true {
			log.Printf("Deleting expired session %s", sessionId)
			s.deleteSession(sessionId)
			return nil
		} else {
			ss.sessionId = sessionId
		}
	}
	return ss
}

func (s *CGISession) Exists(sessionId string) bool {
	s.Driver()
	_, err := s.driver.Retrieve(sessionId)
	if err != nil {
		return false
	}
	return true
}

func (s *CGISession) deleteSession(sessionId string) (bool, error) {
	return s.driver.Remove(sessionId)
}

func (s *CGISession) createSession(sessionId ...string) *SessionStore {
	ss := &SessionStore{session: s}
	if len(sessionId) == 1 && sessionId[0] != "" {
		ss.sessionId = sessionId[0]
	} else {
		ss.sessionId = s.GenerateSessionId()
	}
	now := time.Now().Unix()
	ss.params = make(map[string]interface{})
	ss.ParamInt64("_SESSION_CTIME", now)
	ss.ParamInt64("_SESSION_ATIME", now)
	ss.ParamInt64("_SESSION_ETIME", s.config.ExpireSeconds)
	if ip := os.Getenv("REMOTE_ADDR"); ip != "" {
		ss.ParamString("_SESSION_REMOTE_ADDR", ip)
	}
	ss.ParamString("_SESSION_ID", ss.sessionId)
	ss.Flush()
	return ss
}

func (ss *SessionStore) SessionId() string {
	return ss.sessionId
}

// You must set REMOTE_ADDR environmental variable to make use of IP matching
func (ss *SessionStore) IPMatches() bool {
	// Just return true if we have not enabled IP matching
	if ss.session.config.IPMatch == false {
		log.Printf("IP Checking is disabled")
		return true
	}

	if ip := os.Getenv("REMOTE_ADDR"); ip != "" {
		log.Printf("Checking IP of client %s vs session %s", ip, ss.ParamString("_SESSION_REMOTE_ADDR"))
		if ss.ParamString("_SESSION_REMOTE_ADDR") == ip {
			return true
		}
		log.Printf("THE IP DOES NOT MATCH!\n")
		return false
	} else {
		return true // We don't have the client IP address, so we will not verify it
	}
}

func (ss *SessionStore) ParamInt(name string, value ...int) int {
	if len(value) == 1 {
		ss.params[name] = value[0]
	}
	if ss.params[name] == nil {
		return 0
	}
	return ss.params[name].(int)
}

func (ss *SessionStore) ParamInt64(name string, value ...int64) int64 {
	if len(value) == 1 {
		ss.params[name] = value[0]
	}

	switch ss.params[name].(type) {
	case nil:
		return 0
	case int64:
		return ss.params[name].(int64)
	case string:
		i, _ := strconv.ParseInt(ss.params[name].(string), 10, 64)
		return i
	case float64:
		return int64(ss.params[name].(float64))
	}

	return ss.params[name].(int64)
}

func (ss *SessionStore) GetParam(name interface{}) interface{} {
	if v, ok := ss.params[name.(string)]; ok {
		return v
	}
	return nil
}

func (ss *SessionStore) SetParam(name, value interface{}) error {
	ss.params[name.(string)] = value
	return nil
}

func (ss *SessionStore) ParamFloat64(name string, value ...float64) float64 {
	if len(value) == 1 {
		ss.params[name] = value[0]
	}
	if ss.params[name] == nil {
		return 0
	}
	return ss.params[name].(float64)
}

func (ss *SessionStore) ParamString(name string, value ...string) string {
	if len(value) == 1 {
		ss.params[name] = value[0]
	}
	if ss.params[name] == nil {
		return ""
	}
	return ss.params[name].(string)
}

func (ss *SessionStore) ClearParam(name string) {
	delete(ss.params, name)
}

func (ss *SessionStore) Flush() (bool, error) {
	ss.ParamInt64("_SESSION_ATIME", time.Now().Unix())
	data, err := ss.session.Freeze(ss.params)
	if err != nil {
		return false, err
	}
	_, err = ss.session.driver.Store(ss.sessionId, data)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (ss *SessionStore) IsExpired() bool {
	if (ss.ParamInt64("_SESSION_ATIME") + ss.ParamInt64("_SESSION_ETIME")) <= time.Now().Unix() {
		return true
	}
	return false
}
