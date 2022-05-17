package config

import (
	"context"
	"time"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
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
		return []byte{}, nil
	}

	return resp.Kvs[0].Value, nil
}

func (m *etcdConfigManager) Watch(key string, stop chan bool) <-chan *viper.RemoteResponse {
	viperResponseCh := make(chan *viper.RemoteResponse)

	// need this function to convert the Channel response form crypt.Response to viper.Response
	go func(vr chan<- *viper.RemoteResponse, stop chan bool) {
		client, err := m.client()
		if err != nil {
			log.Logger().Fatal("failed to watch etcd", log.FError(err))
			return
		}

		for {
			ch := client.Watch(context.Background(), key)

			select {
			case <-stop:
				return
			case r := <-ch:
				if r.Canceled {
					return
				}
				if err := r.Err(); err != nil {
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
	}(viperResponseCh, stop)

	return viperResponseCh
}

func (m *etcdConfigManager) client() (*clientv3.Client, error) {
	return clientv3.New(
		clientv3.Config{
			Endpoints:   m.endpoints,
			DialTimeout: time.Second * 5,
			LogConfig: &zap.Config{
				Level:            zap.NewAtomicLevelAt(zap.ErrorLevel),
				Development:      false,
				Encoding:         "json",
				EncoderConfig:    zap.NewProductionEncoderConfig(),
				OutputPaths:      []string{"stderr"},
				ErrorOutputPaths: []string{"stderr"},
			},
		},
	)
}
