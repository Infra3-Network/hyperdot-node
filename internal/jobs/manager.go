package jobs

import (
	"infra-3.xyz/hyperdot-node/internal/store"
	"log"
	"sync/atomic"

	"github.com/jasonlvhit/gocron"
	"infra-3.xyz/hyperdot-node/internal/common"
)

type JobManager struct {
	total          *atomic.Uint64
	cfg            common.Config
	bigquerySyncer *BigQuerySyncer
}

func NewJobManager(cfg *common.Config) *JobManager {
	total := atomic.Uint64{}
	total.Store(0)
	return &JobManager{
		cfg:   *cfg,
		total: &total,
	}
}

func (j *JobManager) Init(boltStore *store.BoltStore) (err error) {
	if j.bigquerySyncer, err = NewBigQuerySyncer(&j.cfg, boltStore); err != nil {
		return
	}

	gocron.Every(1).Day().Do(func() {
		if err := j.bigquerySyncer.Do(); err != nil {
			log.Printf("Error fetching bigquery engine chaindata: %v", err)
			return
		}
	})

	return nil
}

func (j *JobManager) Start() <-chan bool {
	return gocron.Start()
}

func (j *JobManager) Stop() {
	gocron.Clear()
}

// Add every N second job
func (j *JobManager) AddSecondJob(every uint64, name string, job any, args ...any) {
	if every == 1 {
		gocron.Every(1).Second().Do(job, args...)
	} else {
		gocron.Every(every).Seconds().Do(job, args...)
	}
	j.total.Add(1)
	log.Printf("AddSecondJob: %s, total: %d", name, j.total.Load())
}

func (j *JobManager) AddMinuteJob(every uint64, name string, job any, args ...any) {
	if every == 1 {
		gocron.Every(1).Minute().Do(job, args...)
	} else {
		gocron.Every(every).Minutes().Do(job, args...)
	}
	j.total.Add(1)
	log.Printf("AddMinuteJob: %s, total: %d", name, j.total.Load())
}

// Add every N hour job
func (j *JobManager) AddHourJob(every uint64, name string, job any, args ...any) {
	if every == 1 {
		gocron.Every(1).Hour().Do(job, args...)
	} else {
		gocron.Every(every).Hours().Do(job, args...)
	}
	j.total.Add(1)
	log.Printf("AddHourJob: %s, total: %d", name, j.total.Load())
}

// Add every N days job
func (j *JobManager) AddDayJob(every uint64, name string, job any, args ...any) {
	if every == 1 {
		gocron.Every(1).Day().Do(job, args...)
	} else {
		gocron.Every(every).Days().Do(job, args...)
	}
	j.total.Add(1)
	log.Printf("AddDayJob: %s, total: %d", name, j.total.Load())
}
