package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"infra-3.xyz/hyperdot-node/internal/apis/service/query"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

func TestCRUDQuery(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()

	// create
	newQuery := datamodel.QueryModel{
		Name:        "test",
		Description: "test",
		Query:       "select * from `bigquery-public-data.crypto_polkadot.AAA_tableschema` limit 2",
		QueryEngine: "bigquery",
		IsPrivacy:   false,
		Unsaved:     false,
		Stars:       1,
		Charts: []datamodel.ChartModel{
			{Name: "test-1"},
			{Name: "test-2"},
			{Name: "test-3"},
			{Name: "test-4"},
		},
	}

	req, _ := MakeTokenRequest("POST", "/apis/v1/query", newQuery)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	response := query.Response{}
	if err := MarshalResponseBody(w.Body, &response); err != nil {
		t.Fatal(err)
	}

	// get
	w = httptest.NewRecorder()
	req, _ = MakeTokenRequest("GET", fmt.Sprintf("/apis/v1/query/%d", response.Data.ID), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	getResponse := query.Response{}
	if err := MarshalResponseBody(w.Body, &getResponse); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, response.Data.Name, getResponse.Data.Name)

	// update
	updateQuery := datamodel.QueryModel{
		ID:          response.Data.ID,
		Name:        "test-2",
		Description: "test-2",
		Query:       "select * from `bigquery-public-data.crypto_polkadot.AAA_tableschema`",
		QueryEngine: "bigquery",
		IsPrivacy:   false,
		Unsaved:     false,
		Stars:       1,
		Charts: []datamodel.ChartModel{
			{Name: "test-1"},
			{Name: "test-2"},
			{Name: "test-3"},
			{Name: "test-4"},
		},
	}
	w = httptest.NewRecorder()
	req, _ = MakeTokenRequest("PUT", "/apis/v1/query", &updateQuery)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = MakeTokenRequest("GET", fmt.Sprintf("/apis/v1/query/%d", response.Data.ID), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	if err := MarshalResponseBody(w.Body, &getResponse); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, updateQuery.Name, getResponse.Data.Name)

	// delete
	w = httptest.NewRecorder()
	req, _ = MakeTokenRequest("DELETE", fmt.Sprintf("/apis/v1/query/%d", response.Data.ID), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = MakeTokenRequest("GET", fmt.Sprintf("/apis/v1/query/%d", response.Data.ID), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
