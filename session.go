package cgisession

import (
	"strconv"
	"time"

	"github.com/jeffdoubleyou/go-cgi-session/drivers"
	sessionid "github.com/jeffdoubleyou/go-cgi-session/id"
	"github.com/jeffdoubleyou/go-cgi-session/serializers"
)

type CGISessionConfig struct {
	Driver           string
	DriverConfig     interface{}
	Id               string
	IdConfig         interface{}
	Serializer       string
	SerializerConfig interface{}
	ExpireSeconds    int64
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
		sessionConfig = &CGISessionConfig{Driver: "memcached", Id: "md5", Serializer: "datadumper", ExpireSeconds: 86400}
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

func (s *CGISession) New(sessionId ...string) (ss *SessionStore) {
	s.Driver()
	if len(sessionId) > 0 {
		var newSessionId string
		if len(sessionId) == 2 {
			newSessionId = sessionId[1]
		}
		ss = s.Load(sessionId[0])
		if ss == nil {
			ss = s.createSession(newSessionId)
		} else {
			if ss.IsExpired() == true {
				s.deleteSession(sessionId[0])
				ss = s.createSession(newSessionId)
			} else {
				if newSessionId != "" {
					ss = s.createSession(newSessionId)
					s.deleteSession(sessionId[0])
				} else {
					ss.sessionId = sessionId[0]
				}
			}
		}
	} else {
		ss = s.createSession()
	}
	return ss
}

func (s *CGISession) Load(sessionId string) (ss *SessionStore) {
	s.Driver()
	data, err := s.driver.Retrieve(sessionId)
	if err != nil {
		return nil
	} else {
		params, err := s.Thaw(data)
		if err != nil {
			return nil
		}
		if ss.IsExpired() == true {
			s.deleteSession(sessionId)
			return nil
		} else {
			ss.session = s
			ss.sessionId = sessionId
			ss.params = params
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
	if len(sessionId) == 1 {
		ss.sessionId = sessionId[0]
	} else {
		ss.sessionId = s.GenerateSessionId()
	}
	now := time.Now().Unix()
	ss.params = make(map[string]interface{})
	ss.ParamInt64("_SESSION_CTIME", now)
	ss.ParamInt64("_SESSION_ATIME", now)
	ss.ParamInt64("_SESSION_ETIME", s.config.ExpireSeconds)
	ss.ParamString("_SESSION_ID", ss.sessionId)
	ss.Flush()
	return ss
}

func (s *SessionStore) SessionId() string {
	return s.sessionId
}

func (s *SessionStore) ParamInt(name string, value ...int) int {
	if len(value) == 1 {
		s.params[name] = value[0]
	}
	if s.params[name] == nil {
		return 0
	}
	return s.params[name].(int)
}

func (s *SessionStore) ParamInt64(name string, value ...int64) int64 {
	if len(value) == 1 {
		s.params[name] = value[0]
	}

	switch s.params[name].(type) {
	case nil:
		return 0
	case int64:
		return s.params[name].(int64)
	case string:
		i, _ := strconv.ParseInt(s.params[name].(string), 10, 64)
		return i
	case float64:
		return int64(s.params[name].(float64))
	}

	return s.params[name].(int64)
}

func (s *SessionStore) GetParam(name interface{}) interface{} {
	if v, ok := s.params[name.(string)]; ok {
		return v
	}
	return nil
}

func (s *SessionStore) SetParam(name, value interface{}) error {
	s.params[name.(string)] = value
	return nil
}

func (s *SessionStore) ParamFloat64(name string, value ...float64) float64 {
	if len(value) == 1 {
		s.params[name] = value[0]
	}
	if s.params[name] == nil {
		return 0
	}
	return s.params[name].(float64)
}

func (s *SessionStore) ParamString(name string, value ...string) string {
	if len(value) == 1 {
		s.params[name] = value[0]
	}
	if s.params[name] == nil {
		return ""
	}
	return s.params[name].(string)
}

func (s *SessionStore) ClearParam(name string) {
	delete(s.params, name)
}

func (s *SessionStore) Flush() (bool, error) {
	s.ParamInt64("_SESSION_ATIME", time.Now().Unix())
	data, err := s.session.Freeze(s.params)
	if err != nil {
		return false, err
	}
	_, err = s.session.driver.Store(s.sessionId, data)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SessionStore) IsExpired() bool {
	if (s.ParamInt64("_SESSION_ATIME") + s.ParamInt64("_SESSION_ETIME")) <= time.Now().Unix() {
		return true
	}
	return false
}
