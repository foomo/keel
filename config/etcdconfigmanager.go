package config

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type etcdConfigManager struct {
	endpoints []string
}

func NewEtcdConfigManager(endpoints []string) remoteConfigManager {
	return &etcdConfigManager{
		endpoints: endpoints,
	}
}

func (m *etcdConfigManager) Get(key string) ([]byte, error) {
	client, err := m.client()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := client.Get(ctx, key)
	defer cancel()
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
			DialTimeout: 10 * time.Second,
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
			LogConfig: &zap.Config{
				Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
				Development:      false,
				Encoding:         "json",
				EncoderConfig:    zap.NewProductionEncoderConfig(),
				OutputPaths:      []string{"stderr"},
				ErrorOutputPaths: []string{"stderr"},
			},
		},
	)
}
