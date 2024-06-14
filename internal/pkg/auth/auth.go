package auth

import (
	"github.com/GearFramework/gomart/internal/gm/types"
	"github.com/GearFramework/gomart/internal/pkg/alog"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"time"
)

const (
	CookieParamName = "Authorization"
	TokenExpired    = time.Hour * 24
	SecretKey       = "bu&YHU457hgj9Buihiwe7&^jn3iioOOU#J#JJkjjw][]>U#NDW.,ejesf"
)

type CustomerRegisterData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Claims struct {
	jwt.RegisteredClaims
	AuthKey int64
}

type Auth struct {
	logger *zap.SugaredLogger
}

func NewAuth() *Auth {
	return &Auth{
		logger: alog.NewLogger("info"),
	}
}

func (a *Auth) AuthCustomer(ctx *gin.Context) (int64, error) {
	token, err := a.GetTokenFromCookie(ctx)
	if err != nil {
		return 0, err
	}
	return a.getAuthValueFromJWT(token)
}

func (a *Auth) getAuthValueFromJWT(token string) (int64, error) {
	claims, err := a.getClaims(token)
	if err != nil {
		return 0, err
	}
	a.logger.Infof("customer auth value: %d (%T)", claims.AuthKey, claims.AuthKey)
	return claims.AuthKey, nil
}

func (a *Auth) getClaims(tk string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tk, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				a.logger.Errorf("unexpected signing method: %v", t.Header["alg"])
				return nil, types.ErrUnexpectedSigningMethod
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, types.ErrInvalidAuthorization
	}
	return claims, nil
}

func (a *Auth) CreateToken(authValue int64) (string, error) {
	return a.buildJWT(authValue)
}

func (a *Auth) buildJWT(authValue int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpired)),
		},
		AuthKey: authValue,
	})
	tk, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tk, nil
}

func (a *Auth) SetTokenInCookie(ctx *gin.Context, token string) {
	ctx.SetCookie(CookieParamName,
		token,
		int(TokenExpired.Seconds()),
		"/",
		"",
		false,
		true,
	)
}

func (a *Auth) GetTokenFromCookie(ctx *gin.Context) (string, error) {
	c, err := ctx.Request.Cookie(CookieParamName)
	if err != nil {
		return "", types.ErrNeedAuthorization
	}
	if c == nil || c.Value == "" {
		return "", types.ErrInvalidAuthorization
	}
	return c.Value, nil
}
