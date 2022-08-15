package lox

import (
	"net/http"
	"strings"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/rox"
	"github.com/nomos/jwt-go"
	"go.uber.org/zap"
)

const TOKEN_EXPIRE_TIME = time.Minute * 10
const REFRESH_TOKEN_EXPIRE_TIME = time.Hour * 24 * 30

type JwtClaim struct {
	Create time.Time
	Expire int64
}

func NewJwtClaim(expire time.Duration) *JwtClaim {
	return &JwtClaim{
		Create: time.Now(),
		Expire: int64(expire),
	}
}

type JwtClaimWithUser interface {
	jwt.Claims
	GetUser() interface{}
	SetUser(user interface{})
}

type UserCreator func() interface{}

type JwtClaimCreator func(user interface{}, expire time.Duration) JwtClaimWithUser

func (this *JwtClaim) Valid() error {
	expires := this.Create.Add(time.Duration(this.Expire)).Before(time.Now())
	if expires {
		return protocol.ERR_TOKEN_EXPIRED
	}
	return nil
}

func (this *JwtClaim) Marshal() ([]byte, error) {
	return protocol.MarshalBinary(this)
}

func (this *JwtClaim) Unmarshal(v []byte) error {
	return protocol.Unmarshal(v, this)
}

func JwtAuth(rsa bool, creator JwtClaimCreator, userCreator UserCreator) func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	return func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
		w.Header().Set("Access-Control-Expose-Headers", "auth")
		t := r.Header.Get("auth")
		if t == "" {
			t = r.Form.Get("auth")
		}
		log.Warn("auth jwt token", zap.String("token", t))
		if t == "" {
			log.Error("令牌不能为空")
			w.Failed(protocol.ERR_TOKEN_VALIDATE)
			return
		}
		userToken := "Bearer " + t
		split := strings.Split(userToken, " ")
		if len(split) != 2 || split[0] != "Bearer" {
			log.Error("令牌格式不正确")
			w.Failed(protocol.ERR_TOKEN_VALIDATE)
			return
		}
		var key string
		if rsa {
			key = a.Config().GetString("SigningKeyPublic")
		} else {
			key = a.Config().GetString("SigningKey")
		}
		token, err := jwt.ParseWithClaims(split[1], creator(userCreator(), time.Hour*24), func(token *jwt.Token) (interface{}, error) { return []byte(key), nil })
		if err != nil || token.Valid != true {
			// 过期或者非正确处理
			log.Warn("token invalid", zap.String("token", t), zap.Error(err))

			if protocol.ERR_TOKEN_EXPIRED.Is(err.(*jwt.ValidationError).Inner) {
				w.Failed(protocol.ERR_TOKEN_EXPIRED)
			} else {
				w.Failed(protocol.ERR_TOKEN_VALIDATE)
			}
			return
		}
		w.AddContext("token", token.Raw)
		if claim, ok := token.Claims.(JwtClaimWithUser); ok {
			w.AddContext("user", claim.GetUser())
		}
		next.ServeHTTP(w, r)
	}
}

func SignToken(creator JwtClaimCreator, user interface{}, a lokas.IProcess, expire time.Duration, rsa bool) (string, error) {
	var claim *jwt.Token
	var key string
	if rsa {
		key = a.Config().GetString("SigningKeyPrivate")
		claim = jwt.NewWithClaims(jwt.SigningMethodRS256, creator(user, expire))
	} else {
		key = a.Config().GetString("SigningKey")
		claim = jwt.NewWithClaims(jwt.SigningMethodHS256, creator(user, expire))
	}

	token, err := claim.SignedString([]byte(key))
	if err != nil {
		log.Error("sign token failed", zap.Error(err))
		return "", err
	}
	return token, nil
}

func JwtSign(rsa bool, creator JwtClaimCreator) func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	return func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
		next.ServeHTTP(w, r)
		user := w.GetContext("user")
		if user == nil {
			log.Error("用户不存在,令牌无效")
			w.Failed(protocol.ERR_TOKEN_VALIDATE)
			return
		}
		token, err := SignToken(creator, user, a, TOKEN_EXPIRE_TIME, rsa)
		if err != nil {
			log.Error("生成令牌出错")
			w.Response(http.StatusInternalServerError, "服务器繁忙")
			return
		}
		refresh_token, err := SignToken(creator, user, a, REFRESH_TOKEN_EXPIRE_TIME, rsa)
		if err != nil {
			log.Error("生成令牌出错")
			w.Failed(protocol.ERR_TOKEN_VALIDATE)
			return
		}
		w.AddContent("token", token)
		w.AddContent("refresh_token", refresh_token)
	}
}
