package tests

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"infra-3.xyz/hyperdot-node/internal/apis/service/user"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

var apiserver = SetupApiServer()

func TestUserCreateAccount(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	req, _ := MakeTokenRequest("POST", "/apis/v1/user/auth/createAccount", user.RequestCreateAccount{
		Username: fmt.Sprintf("foo-%d", time.Now().Unix()),
		Password: "foo",
		Email:    "foo@email.com",
		Provider: "password",
	})
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUserLogin(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	req, _ := MakeTokenRequest("POST", "/apis/v1/user/auth/login", user.RequestLogin{
		UserId:   "foo",
		Password: "foo",
		Email:    "foo@example.com",
		Provider: "password",
	})
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUserGetCurrent(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	req, _ := MakeTokenRequest("GET", "/apis/v1/user", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUserGet(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	req, _ := MakeTokenRequest("GET", "/apis/v1/user/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUserUpdate(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	username := fmt.Sprintf("bar-%d", time.Now().Unix())
	req, _ := MakeTokenRequest("PUT", "/apis/v1/user", datamodel.UserModel{
		UserBasic: datamodel.UserBasic{
			Username: username,
		},
	})
	router.ServeHTTP(w, req)

	newUser := &user.ResponseUpdateUser{}
	err := MarshalResponseBody(w.Body, newUser)
	assert.Nil(t, err)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, username, newUser.Data.Username)
}

func TestQueryRun(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	path := fmt.Sprintf("/apis/v1/query/run?q=%s&&engine=%s",
		"select * from `bigquery-public-data.crypto_polkadot.AAA_tableschema` limit 2",
		"bigquery")
	req, _ := MakeTokenRequest("GET", path, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
