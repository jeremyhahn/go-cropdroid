package cmd

import (
	"fmt"
	"log"

	"github.com/cockroachdb/pebble"
	"github.com/spf13/cobra"
)

var PebbleDataDir string

func init() {

	pebbleCmd.PersistentFlags().StringVarP(&PebbleDataDir, "data-dir", "", "example-data", "The location of the pebble database directory")

	rootCmd.AddCommand(pebbleCmd)
}

var pebbleCmd = &cobra.Command{
	Use:   "pebble",
	Short: "Manage the pebble database",
	Long:  `Allows managing the embedded pebble database`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("data-dir: %s\n", PebbleDataDir)

		// db, err := pebble.Open(PebbleDataDir, &pebble.Options{})
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// key := []byte("hello")
		// if err := db.Set(key, []byte("world"), pebble.Sync); err != nil {
		// 	log.Fatal(err)
		// }
		// value, closer, err := db.Get(key)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Printf("%s %s\n", key, value)
		// if err := closer.Close(); err != nil {
		// 	log.Fatal(err)
		// }
		// if err := db.Close(); err != nil {
		// 	log.Fatal(err)
		// }

		db, err := pebble.Open(PebbleDataDir, &pebble.Options{})
		if err != nil {
			log.Fatal(err)
		}
		iter := db.NewIter(nil)
		for iter.First(); iter.Valid(); iter.Next() {
			fmt.Printf("%s\n", iter.Key())
		}
		if err := iter.Close(); err != nil {
			log.Fatal(err)
		}
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		// fmt.Println("done")
		// Output:
		// hello
		// hello world
		// world
	},
}
