package index

import (
	"strconv"
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"
)

// Represents a time series record and its related entity timestamp value.
type TimeSeriesIndex struct {
	// "zero byte" prefix to place all time series records first in the table
	// and differentiate between TimeSeries records and entity records. Since
	// entities use non-padded uint64, they should never start with a zero, so
	// this prefix should prevent collisions and provide for fast, deterministic
	// query responses, and constant time lookups between entities and timestamps.
	// When combined with the timestamp, the resulting SSTable will start with all
	//  "zero byte" padded timestamps in the order they were insert into the database,
	// along with the entity ID it corresponds to, followed by all entity values:
	// 1. 1715207765947529:	1
	// 2. 1715207765959604:	2
	// 3. 1:	{"id":1,"name":"Test Entity 1","created":1715207765947529}
	// 4. 2:	{"id":2,"name":"Test Entity 2","created":1715207765959604}
	// 5. applied_index:
	KeyPrefix byte

	// Timestamp as a byte array
	TimestampKey []byte

	// Timestamp value stored in the entity
	Timestamp uint64

	// Entity ID as a byte array
	EntityIDKey []byte

	// Entity ID being indexed
	EntityID uint64
}

// Creates a new fully TimeSeriesIndex with the key prefix, timestamp,
// and timestamp key populated
func NewTimeSeriesIndex() *TimeSeriesIndex {
	index := new(TimeSeriesIndex)
	index.KeyPrefix = byte(0x00)
	index.Timestamp = uint64(time.Now().UnixMicro())
	index.buildTimestampKey()
	return index
}

// Creates a new TimeSeriesIndex with the key prefix, timestamp, and timestamp key
// populated using a provided timestamp.
func CreateTimeSeriesIndex(entity config.TimeSeriesIndexeder) *TimeSeriesIndex {
	index := new(TimeSeriesIndex)
	index.KeyPrefix = byte(0x00)
	index.Timestamp = entity.Timestamp()
	index.EntityID = entity.Identifier()
	index.EntityIDKey = index.encodeUint64(index.EntityID)
	index.buildTimestampKey()
	return index
}

// Sets the timestamp and builds a new timestamp key
func (index *TimeSeriesIndex) SetTimestamp(timestamp uint64) {
	index.Timestamp = timestamp
	index.buildTimestampKey()
}

// Parses a raw key from the database by stripping the KeyPrefix and decoding remaining
// bytes as the TimeSeriesIndex uint64 ID, and sets the Timestamp and TimestampKey.
// An ErrInvalidKeyPrefix error is returned if the TimeSeriesIndex.KeyPrefox is not present.
// When this error is returned while iterating a record set, it means the iterator has gone
// past the time series section of the table, and all time series records have been returned.
func (index *TimeSeriesIndex) ParseKey(key []byte) error {
	prefix, key := key[0], key[1:]
	if prefix != index.KeyPrefix {
		return ErrInvalidKeyPrefix
	}
	ts, err := index.decodeUint64(string(key))
	if err != nil {
		return err
	}
	index.Timestamp = ts
	index.buildTimestampKey()
	return nil
}

// Parses a raw key/value record from the database by stripping the
// prefix and decoding the uint64 timestamp value and referenced entity
// ID and returns a fully populated TimeSeriesIndex with the prefix,
// timestamp, timestamp key, entity id and entity key set. An ErrInvalidKeyPrefix
// error is returned if the TimeSeriesIndex.KeyPrefox is not present. When this error
// is returned while iterating a record set, it means the iterator has gone past the
// time series section of the table, and all time series records have been returned.
func (index *TimeSeriesIndex) ParseKeyValue(key, value []byte) error {
	if err := index.ParseKey(key); err != nil {
		return err
	}
	id, err := index.decodeUint64(string(value))
	if err != nil {
		return err
	}
	index.EntityID = id
	index.EntityIDKey = index.encodeUint64(index.EntityID)
	return nil
}

// Builds the timestamp byte array key
func (index *TimeSeriesIndex) buildTimestampKey() {
	index.TimestampKey = make([]byte, 0)
	index.TimestampKey = append(index.TimestampKey, index.KeyPrefix)
	index.TimestampKey = append(index.TimestampKey, index.encodeUint64(index.Timestamp)...)
}

// Decodes a string literal byte array into a uint64
func (index *TimeSeriesIndex) decodeUint64(value string) (uint64, error) {
	return strconv.ParseUint(string(value), 10, 64)
}

// Encodes a uint64 to a string literal byte array
func (index *TimeSeriesIndex) encodeUint64(id uint64) []byte {
	return []byte(strconv.FormatUint(id, 10))
}
