package store

import (
	bolt "go.etcd.io/bbolt"
	"infra-3.xyz/hyperdot-node/internal/common"
)

type BoltStore struct {
	db *bolt.DB
}

// init bolt
func (b *BoltStore) initDB() error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}

	tx.DB().Batch(func(tx *bolt.Tx) error {
		// metaBucket, err := tx.CreateBucketIfNotExists([]byte("hyperdot_metadata"))
		// if err != nil {
		// 	return err
		// }

		{
			// data, err := json.Marshal([]apis.DataEngine{
			// 	{
			// 		Name: "bigquery",
			// 		Tables: []string{
			// 			"blocks",
			// 			"extrinsic",
			// 		},
			// 	},
			// })
			// if err != nil {
			// 	return err
			// }
			// metaBucket.Put([]byte("engines"), data)

		}

		return nil
	})
	return nil
}

func NewBoltStore(cfg *common.Config) (*BoltStore, error) {
	db, err := bolt.Open(cfg.LocalStore.Bolt.Path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &BoltStore{
		db: db,
	}, nil
}
