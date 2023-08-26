package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"infra-3.xyz/hyperdot-node/internal/apis"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/jobs"
)

var (
	config = flag.String("config", "config.json", "Path to config file")
)

func initialSystemConfig() *common.Config {
	cfg := ""
	if config == nil {
		cfg = os.Getenv("HYPETDOT_NODE_CONFIG")
	} else {
		cfg = *config
	}

	if cfg == "" {
		log.Fatalf("Config file not found")
	}

	configFile, err := os.Open(cfg)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	data, err := io.ReadAll(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	config := new(common.Config)
	if err := json.Unmarshal(data, config); err != nil {
		log.Fatalf("Error unmarshalling config JSON: %v", err)
	}

	return config
}

func initialGlobalData(cfg *common.Config) error {
	// Fetch parachain data and initialize global cache
	if chains, err := jobs.FetchParaChainData(&cfg.Polkaholic); err != nil {
		return err
	} else {
		common.GlobalParaChainCache = common.NewParaChainMap()
		common.GlobalParaChainCache.From(chains)
	}

	return nil
}

func main() {
	flag.Parse()

	cfg := initialSystemConfig()
	if err := initialGlobalData(cfg); err != nil {
		log.Fatalf("Error initial global data: %v", err)
	}

	apiserver := apis.NewApiServer(cfg)
	if err := apiserver.Start(); err != nil {
		log.Fatalf("Error starting api server: %v", err)
	}

	// projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	// if projectID == "" {
	// 	fmt.Println("GOOGLE_CLOUD_PROJECT environment variable must be set.")
	// 	os.Exit(1)
	// }

	// // [START bigquery_simple_app_client]
	// ctx := context.Background()

	// client, err := bigquery.NewClient(ctx, projectID)
	// if err != nil {
	// 	log.Fatalf("bigquery.NewClient: %v", err)
	// }
	// defer client.Close()
	// // [END bigquery_simple_app_client]

	// rows, err := query(ctx, client)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err := printResults(os.Stdout, rows); err != nil {
	// 	log.Fatal(err)
	// }
}

// query returns a row iterator suitable for reading query results.
func query(ctx context.Context, client *bigquery.Client) (*bigquery.RowIterator, error) {

	// [START bigquery_simple_app_query]
	query := client.Query(
		`SELECT
			CONCAT(
				'https://stackoverflow.com/questions/',
				CAST(id as STRING)) as url,
			view_count
		FROM ` + "`bigquery-public-data.stackoverflow.posts_questions`" + `
		WHERE tags like '%google-bigquery%'
		ORDER BY view_count DESC
		LIMIT 10;`)
	return query.Read(ctx)
	// [END bigquery_simple_app_query]
}

// [START bigquery_simple_app_print]
type StackOverflowRow struct {
	URL       string `bigquery:"url"`
	ViewCount int64  `bigquery:"view_count"`
}

// printResults prints results from a query to the Stack Overflow public dataset.
func printResults(w io.Writer, iter *bigquery.RowIterator) error {
	for {
		var row StackOverflowRow
		err := iter.Next(&row)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error iterating through results: %w", err)
		}

		fmt.Fprintf(w, "url: %s views: %d\n", row.URL, row.ViewCount)
	}
}

// [END bigquery_simple_app_print]
// [END bigquery_simple_app_all]
