// +build cluster

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jeremyhahn/cropdroid/state"
	logging "github.com/op/go-logging"
)

type RaftControllerStateDAO struct {
	logger      *logging.Logger
	raftCluster RaftCluster
	datastore.ControllerStateDAO
}

func NewRaftControllerStateDAO(logger *logging.Logger, raftCluster RaftCluster) datastore.ControllerStateDAO {
	return &RaftControllerStateDAO{raftCluster: raftCluster, logger: logger}
}

func (r *RaftControllerStateDAO) Save(controllerID int, controllerState state.ControllerStateMap) error {
	/*
		cs := colfer.ControllerState{}
		bytes, err := cs.MarshalBinary()
		if err != nil {
			panic(err)
		}*/

	r.logger.Errorf("datastore.RaftControllerStateDAO!!! state: %+v", controllerState)

	/*
		clusterID, err := strconv.Atoi(key)
		if err != nil {
			return err
		}*/

	state, err := json.Marshal(controllerState)
	if err != nil {
		return err
	}
	r.raftCluster.SyncPropose(uint64(controllerID), state)
	return nil
	/*
		for k, v := range controllerState.GetMetrics() {
			_, err := r.client.AddAutoTs(fmt.Sprintf("%s_%s", key, k), v)
			if err != nil {
				return err
			}
			println("Saved redis timestamp record")
		}
		return nil
	*/
}

func (r *RaftControllerStateDAO) GetLast30Days(controllerID int, metric string) ([]float64, error) {
	/*
		oneMonthAgo := time.Now().AddDate(0, -1, 0).Unix()
		now := time.Now().Unix()
		datapoints, err := r.client.Range(fmt.Sprintf("%s_%s", key, metric), oneMonthAgo, now)
		floats := make([]float64, len(datapoints))
		for i, datapoint := range datapoints {
			floats[i] = datapoint.Value
		}*/
	var floats []float64
	return floats, nil
}

func (r *RaftControllerStateDAO) createTable(key string, data map[string]float64) error {
	/*
		r.client.CreateKeyWithOptions(key, redistimeseries.DefaultCreateOptions)
		r.client.CreateKeyWithOptions(key+"_avg", redistimeseries.DefaultCreateOptions)
		r.client.CreateRule(key, redistimeseries.MinAggregation, 60, key+"_min")
		r.client.CreateRule(key, redistimeseries.MaxAggregation, 60, key+"_max")
		r.client.CreateRule(key, redistimeseries.AvgAggregation, 60, key+"_avg")
		//return r.client.CreateKeyWithOptions(key, redistimeseries.CreateOptions{RetentionMSecs: 0, Labels: data}) // keep forever
		return r.client.CreateKeyWithOptions(key, redistimeseries.CreateOptions{})
	*/
	return nil
}
