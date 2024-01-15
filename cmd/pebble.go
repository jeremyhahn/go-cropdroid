package cmd

import (
	"encoding/binary"
	"log"
	"strconv"

	"github.com/cockroachdb/pebble"
	"github.com/spf13/cobra"
)

var (
	PebbleGet          string
	PebbleLS           string
	PebbleSetKey       string
	PebbleSetValue     string
	PebbleDecodeUint64 bool
	PebbleEncodeKey    bool
)

func init() {

	pebbleCmd.PersistentFlags().BoolVarP(&PebbleDecodeUint64, "decode-uint64", "", false, "True to decode the returned value as an 8 byte uint64")
	pebbleCmd.PersistentFlags().BoolVarP(&PebbleEncodeKey, "encode-key", "", true, "True to encode the key as a uint64 byte array")

	pebbleCmd.PersistentFlags().StringVarP(&PebbleGet, "get", "", "", "The key to fetch from the database")
	pebbleCmd.PersistentFlags().StringVarP(&PebbleLS, "ls", "", "", "The key to fetch from the database")

	pebbleCmd.PersistentFlags().StringVarP(&PebbleSetKey, "set-key", "", "", "The unique id for the value being stored")
	pebbleCmd.PersistentFlags().StringVarP(&PebbleSetValue, "set-value", "", "", "The value to store")

	rootCmd.AddCommand(pebbleCmd)
}

var pebbleCmd = &cobra.Command{
	Use:   "pebble",
	Short: "Manage the pebble database",
	Long:  `Manages the embedded pebble database`,
	Run: func(cmd *cobra.Command, args []string) {

		db, err := pebble.Open(DataDir, &pebble.Options{})
		if err != nil {
			log.Fatal(err)
		}

		if PebbleSetKey != "" && PebbleSetValue != "" {
			pebbleSet(db, PebbleSetKey, PebbleSetValue)
		} else if PebbleGet != "" {
			pebbleGet(db, PebbleGet)
		} else {
			pebbleLS(db)
		}

		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	},
}

func pebbleLS(db *pebble.DB) {
	i := 0
	iter := db.NewIter(nil)
	for iter.First(); iter.Valid(); iter.Next() {
		i++
		App.Logger.Infof("%d. %s:\t%s\n", i, iter.Key(), iter.Value())
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

func pebbleGet(db *pebble.DB, key string) {
	_key := []byte(key)
	if PebbleEncodeKey {
		uintKey, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			panic(err)
		}
		_key = App.IdGenerator.Uint64Bytes(uintKey)
	}
	val, closer, err := db.Get(_key)
	if err != nil {
		App.Logger.Fatal(err)
	}
	defer closer.Close()
	if len(val) == 0 {
		return
	}
	buf := make([]byte, len(val))
	copy(buf, val)

	if PebbleDecodeUint64 {
		App.Logger.Info(binary.LittleEndian.Uint64(buf))
		return
	}

	App.Logger.Info(string(buf))
}

func pebbleSet(db *pebble.DB, key, value string) {
	options := &pebble.WriteOptions{Sync: true}
	err := db.Set([]byte(key), []byte(value), options)
	if err != nil {
		App.Logger.Fatal(err)
	}
}
