package cluster

import (
	"encoding/json"
	"os"

	"github.com/colin1989/battery/blog"
)

// Server struct
type Server struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Host     string            `json:"host"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Meta     map[string]string `json:"meta"`
	Frontend bool              `json:"frontend"`
	Alive    bool              `json:"alive"`
}

// NewServer ctor
func NewServer(id, serverType, addr string, port int) *Server {
	h, err := os.Hostname()
	if err != nil {
		blog.Error("failed to get hostname: %s", blog.ErrAttr(err))
	}
	return &Server{
		ID:      id,
		Type:    serverType,
		Host:    h,
		Address: addr,
		Port:    port,
	}
}

func NewServerFromBytes(data []byte) (*Server, error) {
	s := Server{}
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Server) GetMeta(name string) (string, bool) {
	if s.Meta == nil {
		return "", false
	}
	v, ok := s.Meta[name]
	return v, ok
}

func (s *Server) SetMeta(name, value string) {
	if s.Meta == nil {
		s.Meta = map[string]string{}
	}
	s.Meta[name] = value
}

// Serialize returns the server as a json string
func (s *Server) Serialize() ([]byte, error) {
	str, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return str, nil
}

func (s *Server) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *Server) IsFrontend() bool {
	return s.Frontend
}

func (s *Server) SetFrontend() {
	s.Frontend = true
}

func (s *Server) IsAlive() bool {
	return s.Alive
}

func (s *Server) SetAlive(alive bool) {
	s.Alive = alive
}

func (s *Server) Equal(other *Server) bool {
	if s == nil || other == nil {
		return false
	}
	if s == other {
		return true
	}
	return s.ID == other.ID
}
