package service

import (
	"fmt"
	"log"
)

type Service struct {
	Name     string
	fullName string
	Metadata map[string]string
	Log      *Logger
}

func New(name string) Service {
	log.SetFlags(0)
	s := Service{
		Name:     name,
		fullName: name,
		Metadata: make(map[string]string),
		Log: &Logger{
			Prefix: fmt.Sprintf("[%s] ", name),
		},
	}
	return s
}

func TestService() Service {
	return New("test")
}

func (s Service) Child(name string) Service {
	fullName := fmt.Sprintf("%s.%s", s.fullName, name)
	log := s.Log.WithPrefix(fmt.Sprintf("[%s] ", fullName))
	return Service{
		Name:     name,
		fullName: fullName,
		Metadata: s.Metadata,
		Log:      &log,
	}
}
