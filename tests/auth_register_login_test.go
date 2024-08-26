package tests

import (
	"fmt"
	suite "sso/tests/suit"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/maximka200/buffpr/gen/go/sso"
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

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{Email: email, Password: password})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

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

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds) // проверка tokenTTL
}

func TestRegisterLogin_DuplicateRegistration(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefLen)

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{Email: email, Password: password})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{Email: email, Password: password})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, fmt.Sprintf("User already exist with email: %s", email))
}

func TestLogin_InvalidCredentials(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefLen)

	// неправильные пароль и логин
	respLog, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{Email: email, Password: password, AppId: 1})
	require.Error(t, err)
	assert.Empty(t, respLog.GetToken())
	assert.ErrorContains(t, err, "Invalid credentials")

	// неправильный пароль или логин
	st.AuthClient.Register(ctx, &ssov1.RegisterRequest{Email: email, Password: password})

	passwordTwo := gofakeit.Password(true, true, true, true, false, passDefLen)
	respLog, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{Email: email, Password: passwordTwo, AppId: 1})
	require.Error(t, err)
	assert.Empty(t, respLog.GetToken())
	assert.ErrorContains(t, err, "Invalid credentials")

	emailTwo := gofakeit.Email()
	respLog, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{Email: emailTwo, Password: password, AppId: 1})
	require.Error(t, err)
	assert.Empty(t, respLog.GetToken())
	assert.ErrorContains(t, err, "Invalid credentials")
}

func TestCreateApp_NewApp_HappyPath(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	name := gofakeit.BeerName()
	secret := gofakeit.BeerName()

	respReg, err := st.AuthClient.CreateApp(ctx, &ssov1.CreateAppRequest{Name: name, Secret: secret})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetAppId())
}

func TestCreateApp_AlreadyExist(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	name := gofakeit.BeerName()
	secret := gofakeit.BeerName()

	respReg, err := st.AuthClient.CreateApp(ctx, &ssov1.CreateAppRequest{Name: name, Secret: secret})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetAppId())

	respReg, err = st.AuthClient.CreateApp(ctx, &ssov1.CreateAppRequest{Name: name, Secret: secret})
	require.Error(t, err)
	assert.Empty(t, respReg.GetAppId())
	assert.ErrorContains(t, err, fmt.Sprintf("App already exist with email: %s", name))
}
