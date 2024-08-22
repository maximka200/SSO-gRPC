package tests

import (
	suite "sso/tests/suit"
	"testing"
	"time"

	ssov1 "github.com/GolangLessons/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppId = 0
	appId      = 1
	appSecret  = "test-secret"

	passDefLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()

	password := gofakeit.Password(true, true, true, true, false, passDefLen)

	respReq, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{Email: email, Password: password})
	require.NoError(t, err)
	assert.NotEmpty(t, respReq.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{Email: email, Password: password, AppId: appId})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReq.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds) // проверка tokenTTL
}
