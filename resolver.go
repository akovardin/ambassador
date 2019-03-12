package ambassador

import (
	"context"

	"google.golang.org/grpc/resolver"
)

// Builder is the implementaion of grpc.naming.Resolver
type Builder struct {
	addr Address
}

// NewBuilder return Builder with service name
func NewBuilder(addr Address) resolver.Builder {
	return &Builder{addr: addr}
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	w, err := NewWatcher(b.addr, target.Endpoint, "", "")
	if err != nil {
		return nil, err
	}

	r := &Resolver{
		conn:    cc,
		watcher: w,
	}

	go w.Run(context.Background())
	go r.watch()

	return r, nil
}

func (b *Builder) Scheme() string {
	return "grpc"
}

type Resolver struct {
	conn    resolver.ClientConn
	watcher *Watcher
}

// It's just a hint, resolver can ignore this if it's not necessary.
func (r *Resolver) ResolveNow(option resolver.ResolveNowOption) {
}

func (r *Resolver) Close() {
	r.watcher.Stop()
}

func (r *Resolver) watch() {
	for aa := range r.watcher.Watch() {
		var addresses []resolver.Address

		for _, a := range aa {
			addresses = append(addresses, resolver.Address{
				Addr: a,
			})
		}

		r.conn.NewAddress(addresses)
	}
}
