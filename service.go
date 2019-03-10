package ambassador

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

type Service struct {
	name   string
	host   string
	port   int
	consul string
	ttl    int
	client *api.Client

	id string
}

func NewService(name string, host string, port int, target string, ttl int) (*Service, error) {
	id := fmt.Sprintf("%s-%s-%d", name, host, port)

	conf := &api.Config{Scheme: "http", Address: target}
	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	return &Service{
		name:   name,
		host:   host,
		port:   port,
		consul: target,
		ttl:    ttl,
		client: client,
		id:     id,
	}, nil
}

func (s *Service) Register() error {
	// initial register service
	regis := &api.AgentServiceRegistration{
		ID:      s.id,
		Name:    s.name,
		Address: s.host,
		Port:    s.port,
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(s.ttl/2))
		for {
			<-ticker.C
			if err := s.client.Agent().UpdateTTL(s.id, "", "passing"); err != nil {
				log.Println("update ttl of service error ", err)
			}
		}
	}()

	err := s.client.Agent().ServiceRegister(regis)
	if err != nil {
		return err
	}

	// initial register service check
	check := api.AgentServiceCheck{TTL: fmt.Sprintf("%ds", s.ttl), Status: "passing"}
	err = s.client.Agent().CheckRegister(&api.AgentCheckRegistration{ID: s.id, Name: s.name, ServiceID: s.id, AgentServiceCheck: check})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Deregister() error {
	err := s.client.Agent().ServiceDeregister(s.id)
	if err != nil {
		return err
	}

	return s.client.Agent().CheckDeregister(s.id)
}
