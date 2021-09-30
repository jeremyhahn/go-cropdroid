package datastore

import (
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/state"

	redistimeseries "github.com/RedisTimeSeries/redistimeseries-go"
)

type RedisClient struct {
	client *redistimeseries.Client
	DeviceDataStore
}

func NewRedisDataStore(address, password string) DeviceDataStore {
	/*
		pool := &redis.Pool{Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address, redis.DialPassword(password))
		}}*/
	return &RedisClient{
		client: redistimeseries.NewClient(":6379", "", nil)}
}

func (r *RedisClient) Client() *redistimeseries.Client {
	return r.client
}

func (r *RedisClient) Save(deviceID uint64, deviceState state.DeviceStateMap) error {
	for k, v := range deviceState.GetMetrics() {
		_, err := r.client.AddAutoTs(fmt.Sprintf("%d_%s", deviceID, k), v)
		if err != nil {
			return err
		}
		println("Saved redis timestamp record")
	}
	for i, v := range deviceState.GetChannels() {
		_, err := r.client.AddAutoTs(fmt.Sprintf("%d_c%d", deviceID, i), float64(v))
		if err != nil {
			return err
		}
		println("Saved redis timestamp record")
	}
	//r.client.MultiAdd(key, data, CreateOptions{RetentionMSecs: 0})
	//values, err = r.tsclient.MultiAdd(Sample{Key: "test:MultiAdd", DataPoint: DataPoint{Timestamp: currentTimestamp + 3, Value: 14}},
	//		Sample{Key: "test:MultiAdd:notExit", DataPoint: DataPoint{Timestamp: currentTimestamp + 4, Value: 14}})
	return nil
}

func (r *RedisClient) GetLast30Days(deviceID uint64, metric string) ([]float64, error) {
	oneMonthAgo := time.Now().AddDate(0, -1, 0).Unix()
	now := time.Now().Unix()
	datapoints, err := r.client.Range(fmt.Sprintf("%d_%s", deviceID, metric), oneMonthAgo, now)
	floats := make([]float64, len(datapoints))
	for i, datapoint := range datapoints {
		floats[i] = datapoint.Value
	}
	return floats, err
}

func (r *RedisClient) createTable(key string, data map[string]float64) error {
	r.client.CreateKeyWithOptions(key, redistimeseries.DefaultCreateOptions)
	r.client.CreateKeyWithOptions(key+"_avg", redistimeseries.DefaultCreateOptions)
	r.client.CreateRule(key, redistimeseries.MinAggregation, 60, key+"_min")
	r.client.CreateRule(key, redistimeseries.MaxAggregation, 60, key+"_max")
	r.client.CreateRule(key, redistimeseries.AvgAggregation, 60, key+"_avg")
	//return r.client.CreateKeyWithOptions(key, redistimeseries.CreateOptions{RetentionMSecs: 0, Labels: data}) // keep forever
	return r.client.CreateKeyWithOptions(key, redistimeseries.CreateOptions{})
}
