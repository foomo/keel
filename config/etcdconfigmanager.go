package config

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type etcdConfigManager struct {
	ctx     context.Context
	client  *clientv3.Client
	timeout time.Duration
}

func NewEtcdConfigManager(endpoints []string) (remoteConfigManager, error) {
	client, err := clientv3.New(
		clientv3.Config{
			Endpoints: endpoints,
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
	if err != nil {
		return nil, err
	}
	return &etcdConfigManager{
		ctx:     context.TODO(),
		client:  client,
		timeout: 3 * time.Second,
	}, nil
}

func (m *etcdConfigManager) Get(key string) ([]byte, error) {
	ctx, cancelFunc := context.WithTimeout(m.ctx, m.timeout)
	defer cancelFunc()

	resp, err := m.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if resp.Count != 1 {
		return nil, fmt.Errorf("getting from etcd with key [%s], res count %d not equal to 1", key, resp.Count)
	}

	return resp.Kvs[0].Value, nil
}

func (m *etcdConfigManager) Watch(key string, stop chan bool) <-chan *viper.RemoteResponse {
	respChan := make(chan *viper.RemoteResponse)
	ctx, cancel := context.WithCancel(m.ctx)
	go func() {
		<-stop
		cancel()
	}()
	// need this function to convert the Channel response form crypt.Response to viper.Response
	go func() {
		for {
			wch := m.client.Watch(ctx, key)

			select {
			case <-ctx.Done():
				fmt.Println("stop watch")
				return
			case we := <-wch:
				for _, event := range we.Events {
					switch event.Type {
					case mvccpb.PUT:
						respChan <- &viper.RemoteResponse{
							Value: event.Kv.Value,
							Error: nil,
						}
					case mvccpb.DELETE:
						// do nothing with delete event
						fmt.Println("find DELETE:", event.PrevKv.Key, event.PrevKv.Value)
					}
				}
			}
		}
	}()

	return respChan
}
