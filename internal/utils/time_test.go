package utils_test

import (
	"testing"
	"time"

	"infra-3.xyz/hyperdot-node/internal/utils"
)

func TestParseTimeRange(t *testing.T) {
	n, d, err := utils.ParseTimeRange("7h")
	if err != nil {
		t.Fatal(err)
	}

	if n != 7 || d != time.Hour {
		t.Fatal("invalid time range")
	}

	n, d, err = utils.ParseTimeRange("24h")
	if err != nil {
		t.Fatal(err)
	}

	if n != 24 || d != time.Hour {
		t.Fatal("invalid time range")
	}

	n, d, err = utils.ParseTimeRange("7d")
	if err != nil {
		t.Fatal(err)
	}

	if n != 7 || d != time.Hour*24 {
		t.Fatal("invalid time range")
	}

	n, d, err = utils.ParseTimeRange("24d")
	if err != nil {
		t.Fatal(err)
	}

	if n != 24 || d != time.Hour*24 {
		t.Fatal("invalid time range")
	}

}

func TestGetTimeBefore(t *testing.T) {
	now := time.Now()
	t1 := utils.GetTimeBefore(7)
	if t1.After(now) {
		t.Fatal("invalid time")
	}

	t2 := utils.GetTimeBeforeDays(7)
	if t2.After(now) {
		t.Fatal("invalid time")
	}
}
