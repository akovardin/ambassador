package ambassador

import (
	"context"
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
)

type Watcher struct {
	service   string
	plan      *watch.Plan
	addresses chan []string
	consul    string
}

func NewWatcher(consul, service, dc, tag string) (*Watcher, error) {
	plan, err := watch.Parse(map[string]interface{}{
		"type":        "service",
		"service":     service,
		"datacenter":  dc,
		"tag":         tag,
		"passingonly": true,
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan []string)
	plan.Handler = func(u uint64, data interface{}) {
		var addrs []string
		for _, srv := range data.([]*api.ServiceEntry) {
			host := srv.Service.Address
			if host == "" {
				host = srv.Node.Address
			}
			addrs = append(addrs, fmt.Sprintf("%s:%d", host, srv.Service.Port))
		}
		ch <- addrs
	}

	return &Watcher{
		service:   service,
		plan:      plan,
		addresses: ch,
		consul:    consul,
	}, nil
}

func (w *Watcher) Watch() chan []string {
	return w.addresses
}

func (w *Watcher) Run(ctx context.Context) error {
	return w.plan.Run(w.consul)
}

func (w *Watcher) Stop() {
	w.plan.Stop()
}
