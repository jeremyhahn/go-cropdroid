package cockroach

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jinzhu/gorm"
)

type CockroachChangefeed struct {
	app   *app.App
	db    *gorm.DB
	table string
	datastore.Changefeeder
}

type changefeedWrapper struct {
	After   interface{} `json:"after"`
	Updated string      `json:"updated"`
}

type changefeedRawMessageWrapper struct {
	After   map[string]*json.RawMessage `json:"after"`
	Updated string                      `json:"updated"`
}

func NewCockroachChangefeed(app *app.App, table string) datastore.Changefeeder {
	return &CockroachChangefeed{
		app:   app,
		db:    app.NewGormDB(),
		table: table}
}

func (changefeed *CockroachChangefeed) Subscribe(callback datastore.ChangefeedCallback) {

	var key string
	var value string
	sleepTime := 30 * time.Second

	sleepAndReSubscribe := func() {
		changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error: Unable to connect to database. Retrying again in %d seconds", sleepTime)
		time.Sleep(sleepTime)
		go changefeed.Subscribe(callback)
	}

	changefeed.app.Logger.Debugf("[CockroachChangefeed.Subscribe] table=%s", changefeed.table)

	rows, err := changefeed.db.Raw(fmt.Sprintf("EXPERIMENTAL CHANGEFEED FOR %s WITH updated;", changefeed.table)).Rows()
	if err != nil {
		if err.Error() == fmt.Sprintf("pq: table \"%s\" does not exist", changefeed.table) {
			changefeed.app.Logger.Warningf("CockroachChangefeed table not found: %s", err)
			sleepAndReSubscribe()
			return
		}
		if err.Error() == "server is not accepting clients" {
			sleepAndReSubscribe()
			return
		}
		changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error: %s", err)
	}

	if rows == nil {
		sleepAndReSubscribe()
		return
	}

	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&changefeed.table, &key, &value)
		if err != nil {
			changefeed.app.Logger.Fatal(err)
		}

		go func(table, key, value string) {
			changefeed.app.Logger.Warningf("[CockroachChangefeed.Subscribe] result: table=%s, key=%s, value=%s", changefeed.table, key, value)
			key = strings.Replace(key, "[", "", 1)
			key = strings.Replace(key, "]", "", 1)
			int64Key, err := strconv.ParseInt(key, 0, 64)
			if err != nil {
				changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error parsing int64 table primary key: ", err)
				return
			}
			/*
				if strings.Contains(table, "_state") {
					// controller state table
					var rawMessageWrapper changefeedRawMessageWrapper
					if err = json.Unmarshal([]byte(value), &rawMessageWrapper); err != nil {
						changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error unmarshaling into changefeedStateWrapper: %s", err)
						return
					}
					callback(&datastore.DatabaseChangefeed{
						Table: table,
						Key:   int64Key,
						//Bytes: jsonBytes,
						Updated:    wrapper.Updated,
						RawMessage: wrapper.After})
					return
				}
				// cropdroid system table
				var wrapper changefeedWrapper
				if err = json.Unmarshal([]byte(value), &wrapper); err != nil {
					changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error unmarshaling into changefeedWrapper: %s", err)
					return
				}
				jsonBytes, err := json.Marshal(wrapper.After)
				if err != nil {
					changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error unmarshaling map to string: %s", err)
					return
				}
				callback(&datastore.DatabaseChangefeed{
					Table: table,
					Key:   int64Key,
					//Value:   wrapper.After,
					Updated: wrapper.Updated,
					Bytes:   jsonBytes})
			*/
			var rawMessageWrapper changefeedRawMessageWrapper
			if err = json.Unmarshal([]byte(value), &rawMessageWrapper); err != nil {
				changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error unmarshaling into changefeedStateWrapper: %s", err)
				return
			}
			var wrapper changefeedWrapper
			if err = json.Unmarshal([]byte(value), &wrapper); err != nil {
				changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error unmarshaling into changefeedWrapper: %s", err)
				return
			}
			jsonBytes, err := json.Marshal(wrapper.After)
			if err != nil {
				changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Error unmarshaling map to string: %s", err)
				return
			}
			callback(&datastore.DatabaseChangefeed{
				Table: table,
				Key:   int64Key,
				//Value:   wrapper.After,
				Updated:    wrapper.Updated,
				Bytes:      jsonBytes,
				RawMessage: rawMessageWrapper.After})
		}(changefeed.table, key, value)
	}

	changefeed.app.Logger.Errorf("[CockroachChangefeed.Subscribe] Closing changefeed: table=%s", changefeed.table)
	sleepAndReSubscribe()
}
