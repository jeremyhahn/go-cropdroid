package index

import (
	"strconv"

	"github.com/jeremyhahn/go-cropdroid/config"
)

// Represents a time series record and its related entity timestamp value.
type Reference struct {

	// Unique key prefix for this referennce, ex: []byte{"o", ":"}
	KeyPrefix []byte

	// Root entity ID as a byte array
	RootKey []byte

	// Root entity ID value stored in the entity
	RootID uint64

	// Referenced entity ID as a byte array
	RefIDKey []byte

	// Referenced entity ID being indexed
	RefID uint64
}

func NewReference(prefix []byte, rootID uint64) *Reference {
	index := new(Reference)
	index.KeyPrefix = prefix
	index.RootID = rootID
	index.buildRootKey()
	return index
}

func CreateReference(KeyPrefix []byte, entity config.TimeSeriesIndexeder) *TimeSeriesIndex {
	index := new(TimeSeriesIndex)
	index.KeyPrefix = byte(0x01)
	index.Timestamp = entity.Timestamp()
	index.EntityID = entity.Identifier()
	index.EntityIDKey = index.encodeUint64(index.EntityID)
	index.buildTimestampKey()
	return index
}

// Parses a raw key from the database by stripping the KeyPrefix and decoding remaining
// bytes as the TimeSeriesIndex uint64 ID, and sets the Timestamp and TimestampKey.
// An ErrInvalidKeyPrefix error is returned if the TimeSeriesIndex.KeyPrefox is not present.
// When this error is returned while iterating a record set, it means the iterator has gone
// past the time series section of the table, and all time series records have been returned.
func (ref *Reference) ParseKey(key []byte) error {

	prefix1, key := key[0], key[1:]
	prefix2, key := key[0], key[1:]
	if prefix1 != ref.KeyPrefix[0] && prefix2 != ref.KeyPrefix[1] {
		return ErrInvalidKeyPrefix
	}

	rootID, err := ref.decodeUint64(string(key))
	if err != nil {
		return err
	}

	ref.RootID = rootID
	ref.buildRootKey()

	return nil
}

// Parses a raw key/value record from the database by stripping the
// prefix and decoding the uint64 timestamp value and referenced entity
// ID and returns a fully populated TimeSeriesIndex with the prefix,
// timestamp, timestamp key, entity id and entity key set. An ErrInvalidKeyPrefix
// error is returned if the TimeSeriesIndex.KeyPrefox is not present. When this error
// is returned while iterating a record set, it means the iterator has gone past the
// time series section of the table, and all time series records have been returned.
func (ref *Reference) ParseKeyValue(key, value []byte) error {
	if err := ref.ParseKey(key); err != nil {
		return err
	}
	id, err := ref.decodeUint64(string(value))
	if err != nil {
		return err
	}
	ref.RefID = id
	ref.RefIDKey = ref.encodeUint64(ref.RefID)
	return nil
}

// Builds a root key byte array using the prefix and root ID
func (ref *Reference) buildRootKey() {
	ref.RootKey = make([]byte, 0)
	ref.RootKey = append(ref.RootKey, ref.KeyPrefix...)
	ref.RootKey = append(ref.RootKey, ref.encodeUint64(ref.RootID)...)
}

// Decodes a string literal byte array into a uint64
func (ref *Reference) decodeUint64(value string) (uint64, error) {
	return strconv.ParseUint(string(value), 10, 64)
}

// Encodes a uint64 to a string literal byte array
func (ref *Reference) encodeUint64(id uint64) []byte {
	return []byte(strconv.FormatUint(id, 10))
}
