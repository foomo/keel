package config

import (
	"context"
	"time"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type etcdConfigManager struct {
	endpoints []string
	l         *zap.Logger
}

func NewEtcdConfigManager(l *zap.Logger, endpoints []string) remoteConfigManager {
	return &etcdConfigManager{
		endpoints: endpoints,
		l:         l,
	}
}

func (m *etcdConfigManager) Get(key string) ([]byte, error) {
	client, err := m.client()
	if err != nil {
		return nil, err
	}
	defer func(client *clientv3.Client) {
		_ = client.Close()
	}(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := client.Get(ctx, key)
	cancel()
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0].Value, nil
}

// TODO sth is broken ok re-connect
func (m *etcdConfigManager) Watch(key string, stop chan bool) <-chan *viper.RemoteResponse {
	viperResponsCh := make(chan *viper.RemoteResponse)
	client, err := m.client()
	if err != nil {
		viperResponsCh <- &viper.RemoteResponse{
			Value: nil,
			Error: err,
		}
		return viperResponsCh
	}

	watcher := clientv3.NewWatcher(client)
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	watchRespChan := watcher.Watch(cancelCtx, key)
	// need this function to convert the Channel response form crypt.Response to viper.Response
	go func(cr clientv3.WatchChan, vr chan<- *viper.RemoteResponse, cancelFunc context.CancelFunc, stop chan bool) {
		for {
			select {
			case <-stop:
				cancelFunc()
				return
			case r := <-cr:
				if r.Canceled {
					cancelFunc()
					return
				}
				if err := r.Err(); err != nil {
					vr <- &viper.RemoteResponse{
						Value: nil,
						Error: err,
					}
					continue
				}
				for _, event := range r.Events {
					vr <- &viper.RemoteResponse{
						Value: event.Kv.Value,
						Error: nil,
					}
				}
			}
		}
	}(watchRespChan, viperResponsCh, cancelFunc, stop)

	return viperResponsCh
}

func (m *etcdConfigManager) client() (*clientv3.Client, error) {
	return clientv3.New(
		clientv3.Config{
			Endpoints:   m.endpoints,
			DialTimeout: time.Second,
			DialOptions: []grpc.DialOption{
				grpc.WithBlock(),
				grpc.WithDefaultCallOptions(
					grpc.WaitForReady(true),
				),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			},
			Logger: m.l,
		},
	)
}
