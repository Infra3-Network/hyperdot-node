package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

	// hack to skip test if user already exists
	if w.Code == http.StatusBadRequest &&
		strings.Contains(w.Body.String(), "already exists") {
		return
	}
	assert.Equal(t, 200, w.Code)
}

func TestUserLogin(t *testing.T) {
	router := apiserver.GetEngine()
	w := httptest.NewRecorder()
	req, _ := MakeTokenRequest("POST", "/apis/v1/user/auth/login", user.RequestLogin{
		Password: "foo",
		Email:    "foo@email.com",
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
	bio := "test update user"
	req, _ := MakeTokenRequest("PUT", "/apis/v1/user", datamodel.UserModel{
		UserBasic: datamodel.UserBasic{
			Bio: bio,
		},
	})
	router.ServeHTTP(w, req)

	newUser := &user.ResponseUpdateUser{}
	err := MarshalResponseBody(w.Body, newUser)
	assert.Nil(t, err)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, bio, newUser.Data.Bio)
}
