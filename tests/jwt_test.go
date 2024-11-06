package tests

import (
	"github.com/brianvoe/gofakeit"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"sso/internal/domain/models"
	jwtlocal "sso/internal/lib"
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {

	app := models.App{
		Id:     gofakeit.Number(1, 100),
		Name:   gofakeit.Name(),
		Secret: []byte(gofakeit.Name()),
	}

	passwordUsr := gofakeit.Password(
		true, true, true, true, true, 10)

	passHash, err := bcrypt.GenerateFromPassword([]byte(passwordUsr), bcrypt.DefaultCost)
	assert.NoError(t, err)

	roles := []string{"admin", "user"}

	usr := models.User{
		ID:       int64(gofakeit.Number(1, 100)),
		Email:    gofakeit.Email(),
		PassHash: passHash,
		Roles:    roles,
	}

	ttl := time.Hour

	token, err := jwtlocal.NewToken(usr, app, ttl)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	decodedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return app.Secret, nil
	})
	assert.NoError(t, err)

	assert.True(t, decodedToken.Valid)

	claims, ok := decodedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	/*
		claims["uid"] = user.ID
		claims["email"] = user.Email
		claims["app_id"] = app.Id
		claims["roles"] = user.Roles
	*/
	rolesJWT, ok := claims["roles"].([]interface{})
	if !ok {
		t.Error("cannot assert type")
	}
	actualRoles := make([]string, len(rolesJWT))
	for i, v := range rolesJWT {
		actualRoles[i] = v.(string)
	}
	assert.Equal(t, roles, actualRoles)
	assert.Equal(t, usr.Email, claims["email"])
	assert.Equal(t, float64(app.Id), claims["app_id"])
}
