package store

import (
	"encoding/json"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"

	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

const (
	METADATA_BUCKET = "hyperdot_chain_metadata"
)

type BoltStore struct {
	db *bolt.DB
}

// init bolt
func (b *BoltStore) initDB() error {
	log.Printf("initDB")
	engines := []datamodel.QueryEngine{
		{
			Name: "Bigquery",
			Datasets: map[string]datamodel.QueryEngineDatasetMetadata{
				"Bigquery": {
					Id:          "raw",
					Title:       "Raw",
					Description: "Raw blockchain crypto data",
				},
			},
		},
	}

	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}

	metaBucket, err := tx.CreateBucketIfNotExists([]byte("hyperdot_chain_metadata"))
	if err != nil {
		log.Printf("initDB: %v", err)
		return err
	}

	jsonData, err := json.Marshal(engines)
	if err != nil {
		log.Printf("initDB: %v", err)
		return err
	}

	log.Printf("initDB: %v", string(jsonData))

	if err := metaBucket.Put([]byte("query_engines"), jsonData); err != nil {
		log.Printf("initDB: %v", err)
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) SetDatasets(queryEngine string, data *datamodel.QueryEngineDatasets) error {
	log.Printf("SetBigQueryChainData: %v", data)
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}

	metaBucket, err := tx.CreateBucketIfNotExists([]byte("hyperdot_chain_metadata"))
	if data.Raw != nil {
		key := fmt.Sprintf("%s_%s", queryEngine, "raw")
		jsonData, err := json.Marshal(data.Raw)
		if err != nil {
			return err
		}

		if err := metaBucket.Put([]byte(key), jsonData); err != nil {
			return tx.Rollback()
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (b *BoltStore) GetDataset(queryEngine string, tag string) (*datamodel.QueryEngineDatasetInfo, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}

	metaBucket := tx.Bucket([]byte(METADATA_BUCKET))
	if metaBucket == nil {
		return nil, fmt.Errorf("bucket %s not found", METADATA_BUCKET)
	}

	key := fmt.Sprintf("%s_%s", queryEngine, tag)
	data := metaBucket.Get([]byte(key))
	if data == nil {
		return nil, fmt.Errorf("%s of %s not found", queryEngine, tag)
	}

	bigqueryData := new(datamodel.QueryEngineDatasetInfo)
	if err := json.Unmarshal(data, bigqueryData); err != nil {
		return nil, err
	}

	return bigqueryData, nil
}

func (b *BoltStore) GetQueryEngines() ([]datamodel.QueryEngine, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}

	metaBucket := tx.Bucket([]byte("hyperdot_chain_metadata"))
	if metaBucket == nil {
		return nil, nil
	}

	data := metaBucket.Get([]byte("query_engines"))
	if data == nil {
		return nil, nil
	}

	queryEngines := make([]datamodel.QueryEngine, 0)
	if err := json.Unmarshal(data, &queryEngines); err != nil {
		return nil, err
	}

	return queryEngines, nil
}

func NewBoltStore(cfg *common.Config) (*BoltStore, error) {
	db, err := bolt.Open(cfg.LocalStore.Bolt.Path, 0600, nil)
	if err != nil {
		return nil, err
	}
	b := &BoltStore{
		db: db,
	}

	if err := b.initDB(); err != nil {
		return nil, err
	}

	return b, nil
}
