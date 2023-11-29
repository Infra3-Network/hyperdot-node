package jobs

import (
	"log"
	"sync/atomic"

	"infra-3.xyz/hyperdot-node/internal/store"

	"github.com/jasonlvhit/gocron"
	"infra-3.xyz/hyperdot-node/internal/common"
)

// JobManager is controller of all jobs and can timing execute jobs
type JobManager struct {
	total          *atomic.Uint64
	cfg            common.Config
	bigquerySyncer *BigQuerySyncer
}

// NewJobManager creates a new JobManager
func NewJobManager(cfg *common.Config) *JobManager {
	total := atomic.Uint64{}
	total.Store(0)
	return &JobManager{
		cfg:   *cfg,
		total: &total,
	}
}

// Init initializes the job manager
// It starts theses jobs
//  1. bigquery syncer
func (j *JobManager) Init(boltStore *store.BoltStore) (err error) {
	if j.bigquerySyncer, err = NewBigQuerySyncer(&j.cfg, boltStore); err != nil {
		return
	}

	err = gocron.Every(1).Day().From(gocron.NextTick()).Do(func() {
		if err := j.bigquerySyncer.Do(); err != nil {
			log.Printf("Error fetching bigquery engine chaindata: %v", err)
			return
		}
	})

	return err
}

// Start starts the job manager
func (j *JobManager) Start() <-chan bool {
	return gocron.Start()
}

// Stop stops the job manager
func (j *JobManager) Stop() {
	gocron.Clear()
}
