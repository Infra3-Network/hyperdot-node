package tests

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

func TestDashboardCreate(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	body := datamodel.DashboardModel{
		Name:        "test",
		Description: "test",
		IsPrivacy:   false,
		Panels: []datamodel.DashboardPanelModel{
			{Name: "test-1"},
			{Name: "test-2"},
			{Name: "test-3"},
		},
	}
	req, _ := MakeTokenRequest("POST", "/apis/v1/dashboard", body)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
