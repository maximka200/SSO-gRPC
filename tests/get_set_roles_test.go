package tests

import (
	"context"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"sso/internal/config"
	storage "sso/internal/storage/postgresql"
	"testing"
)

func TestGetSetRoles(t *testing.T) {

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefLen)
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	cfg := config.MustByLoad("../config/localv2.yaml")
	db, err := storage.NewDB(cfg)
	assert.NoError(t, err)

	db.SaveUser(context.Background(), email, passHash)
	roles := []string{"admin", "user"}

	err = db.SetRoles(context.Background(), email, roles)
	assert.NoError(t, err)

	rolesGet, err := db.GetRoles(context.Background(), email)
	assert.NoError(t, err)

	assert.Equal(t, roles, rolesGet)
}
