package tests

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"infra-3.xyz/hyperdot-node/internal/apis/service/user"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

var apiserver = SetupApiServer()

func TestUserCreateAccount(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	req, _ := MakeTokenRequest("POST", "/apis/v1/user/auth/createAccount", user.RequestCreateAccount{
		Username: "foo",
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
	req, _ := MakeTokenRequest("PUT", "/apis/v1/user", datamodel.UserModel{
		UserBasic: datamodel.UserBasic{
			Username: "bar",
		},
	})
	router.ServeHTTP(w, req)

	newUser := &user.ResponseUpdateUser{}
	err := MarshalResponseBody(w.Body, newUser)
	assert.Nil(t, err)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "bar", newUser.Data.Username)
}
