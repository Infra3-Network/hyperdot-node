package jobs_test

import (
	"testing"

	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/jobs"
)

func TestChainJob(t *testing.T) {
	cfg := common.Config{
		Polkaholic: common.PolkaholicConfig{
			ApiKey:  "aaed8e0afefcf294e146167fbca9814a",
			BaseUrl: "https://api.polkaholic.io",
		},
		Bigquery: common.BigQueryConfig{
			// ProjectId: "substarte-etl",
			ProjectId: "substarte-etl",
		},
	}

	job, err := jobs.NewBigQuerySyncer(&cfg)
	if err != nil {
		t.Fatalf("Error creating parachain job: %v", err)
	}

	_, err = job.Do()
	if err != nil {
		t.Logf("Error fetching parachain data: %v", err)
	}
}
