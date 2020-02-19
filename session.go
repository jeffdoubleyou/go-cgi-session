package cgisession

import (
    "time"
    "strconv"
    "go-cgi-session/drivers"
    "go-cgi-session/id"
    "go-cgi-session/serializers"
)

type config struct {
    driver string
    driverConfig map[string]interface{}
    id string
    idConfig map[string]interface{}
    serializer string
    serializerConfig map[string]interface{}
    expireSeconds int64
    cookieName string
    cookieDomain string
    cookiePath string
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
    config *config
    driver driver
    id id
    serializer serializer
    sessionId string
    params map[string]interface{}
}

func Session(props ...map[string]interface{}) *CGISession {
    sessionConfig := &config{driver: "memcached", id: "md5", serializer: "datadumper", expireSeconds: 86400, cookieName: "CGISESSID", cookiePath: "/"}
    if len(props) > 0 {
        if props[0]["driverConfig"] != nil {
            sessionConfig.driverConfig = props[0]["driverConfig"].(map[string]interface{})
        }
        if props[0]["idConfig"] != nil {
            sessionConfig.idConfig = props[0]["idConfig"].(map[string]interface{})
        }
        if c := props[0]["cookieName"]; c != nil {
            sessionConfig.cookieName = c.(string)
        }
        if c := props[0]["cookieDomain"]; c != nil {
            sessionConfig.cookieDomain = c.(string)
        }
        if c := props[0]["cookiePath"]; c != nil {
            sessionConfig.cookiePath = c.(string)
        }
    }
    return &CGISession{config: sessionConfig}
}

func (s *CGISession) Id(idgen ...id) {
    if s.id == nil {
        if len(idgen) == 1 {
            s.id = idgen[0]
        } else {
            switch idGen := s.config.id; idGen {
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
            switch ss := s.config.serializer; ss {
                case "datadumper":
                    s.serializer = serializers.DataDumper(s.config.serializerConfig)
                default:
                    s.serializer = serializers.DataDumper(s.config.serializerConfig)
            }
        }
    }
}

func (s *CGISession) Thaw(data []byte) (bool, error) {
    s.Serializer()
    params, err := s.serializer.Thaw(data)
    if err != nil {
        return false, err
    }
    s.params = params.(map[string]interface{})
    return true, nil
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
            switch d := s.config.driver; d {
                case "memcached":
                    s.driver = drivers.Memcached(s.config.driverConfig)
                default:
                    s.driver = drivers.Memcached(s.config.driverConfig)
            }
        }
    }
}

func (s *CGISession) New(sessionId ...string) interface{} {
    s.Driver()
    if len(sessionId) > 0 {
        var newSessionId string
        if len(sessionId) == 2 {
            newSessionId = sessionId[1]
        }
        data, err := s.driver.Retrieve(sessionId[0])
        if err != nil {
            s.createSession(newSessionId)
        } else {
            _, err := s.Thaw(data)
            if err != nil {
                s.createSession(newSessionId)
            }
            if s.IsExpired() == true {
                s.deleteSession(sessionId[0])
                s.createSession(newSessionId)
            } else {
                if newSessionId != "" {
                    s.createSession(newSessionId)
                    s.deleteSession(sessionId[0])
                    s.Flush()
                } else {
                    s.sessionId = sessionId[0]
                }
            }
        }
    } else {
        s.createSession()
    }
    return s.params 
}

func (s *CGISession) Load(sessionId string) interface{} {
    s.Driver()
    data, err := s.driver.Retrieve(sessionId)
    if err != nil {
        return nil
    } else {
        _, err := s.Thaw(data)
        if err != nil {
            return nil
        }
        if s.IsExpired() == true {
            s.deleteSession(sessionId)
            return nil
        } else {
            s.sessionId = sessionId
        }
    }
    return s.params 
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

func (s *CGISession) createSession(sessionId ...string) string {
    if len(sessionId) == 1 {
        s.sessionId = sessionId[0]
    } else {
        s.sessionId = s.GenerateSessionId()
    }
    now := time.Now().Unix()
    s.params = make(map[string]interface{})
    s.ParamInt64("_SESSION_CTIME", now)
    s.ParamInt64("_SESSION_ATIME", now)
    s.ParamInt64("_SESSION_ETIME", s.config.expireSeconds)
    s.ParamString("_SESSION_ID", s.sessionId)
    return s.sessionId
}

func (s *CGISession) SessionId() string {
    return s.sessionId
}

func (s *CGISession) ParamInt(name string, value ...int) int {
    if len(value) == 1 {
        s.params[name] = value[0]
    }
    return s.params[name].(int)
}

func (s *CGISession) ParamInt64(name string, value ...int64) int64 {
    if len(value) == 1 {
        s.params[name] = value[0]
    }

    switch s.params[name].(type) {
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

func (s *CGISession) GetParam(name interface{}) interface{} {
    if v, ok := s.params[name.(string)]; ok {
        return v
    }
    return nil
}

func (s *CGISession) SetParam(name, value interface{}) error {
    s.params[name.(string)] = value
    return nil
}

func (s *CGISession) ParamFloat64(name string, value ...float64) float64 {
    if len(value) == 1 {
        s.params[name] = value[0]
    }
    return s.params[name].(float64)
}

func (s *CGISession) ParamString(name string, value ...string) string {
    if len(value) == 1 {
        s.params[name] = value[0]
    }
    return s.params[name].(string)
}

func (s *CGISession) ClearParam(name string) {
    delete(s.params, name)
}

func (s *CGISession) Flush() (bool, error) {
    s.ParamInt64("_SESSION_ATIME", time.Now().Unix())
    data, err := s.Freeze(s.params)
    if err != nil {
        return false, err
    }
    _, err = s.driver.Store(s.sessionId, data)
    if err != nil {
        return false, err
    }
    return true, nil
}

func (s *CGISession) IsExpired() bool {
    if (s.ParamInt64("_SESSION_ATIME") + s.ParamInt64("_SESSION_ETIME")) <= time.Now().Unix() {
        return true
    }
    return false
}

