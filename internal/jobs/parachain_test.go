package jobs_test

import (
	"testing"

	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/jobs"
)

func TestParaChainJob(t *testing.T) {
	cfg := common.Config{
		Pokaholic: common.Polkaholic{
			ApiKey:  "aaed8e0afefcf294e146167fbca9814a",
			BaseUrl: "https://api.polkaholic.io",
		},
	}

	job := jobs.NewFetchParaChain(&cfg)
	chains, err := job.Do()
	if err != nil {
		t.Logf("Error fetching parachain data: %v", err)
	}

	if len(chains) == 0 {
		t.Error("No parachain data returned")
	}
}
