package lox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/lox/rox"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/jwt-go"
	"net/http"
	"strings"
	"time"
)

const TOKEN_EXPIRE_TIME = time.Minute*10
const REFRESH_TOKEN_EXPIRE_TIME = time.Hour*24*30

type JwtClaim struct {
	Create time.Time
	Expire int64
}

func NewJwtClaim(expire time.Duration)*JwtClaim {
	return &JwtClaim{
		Create: time.Now(),
		Expire: int64(expire),
	}
}

type JwtClaimWithUser interface {
	jwt.Claims
	GetUser()interface{}
	SetUser(user interface{})
}

type UserCreator func()interface{}

type JwtClaimCreator func(user interface{},expire time.Duration)JwtClaimWithUser

func (this* JwtClaim) Valid() error {
	expires:= this.Create.Add(time.Duration(this.Expire)).Before(time.Now())
	if expires {
		return protocol.ErrTokenExpired
	}
	return nil
}

func (this* JwtClaim) Marshal()([]byte,error) {
	return protocol.MarshalBinary(this)
}

func (this* JwtClaim) Unmarshal(v []byte)error {
	return protocol.Unmarshal(v,this)
}


func JwtAuth(rsa bool,creator JwtClaimCreator,userCreator UserCreator)func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	return func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
		w.Header().Set("Access-Control-Expose-Headers", "auth")
		t:=r.Header.Get("auth")
		if t=="" {
			t = r.Form.Get("auth")
		}
		log.Warnf("token",t)
		if t == "" {
			log.Error("令牌不能为空")
			w.Failed(protocol.ErrTokenValidate)
			return
		}
		userToken := "Bearer "+t
		split := strings.Split(userToken, " ")
		if len(split) != 2 || split[0] != "Bearer"{
			log.Error("令牌格式不正确")
			w.Failed(protocol.ErrTokenValidate)
			return
		}
		var key interface{}
		if rsa {
			key=a.Config().Get("SigningKeyPublic")
		} else {
			key = a.Config().Get("SigningKey")
		}
		token, err := jwt.ParseWithClaims(split[1], creator(userCreator(),time.Hour*24), func(token *jwt.Token) (interface{}, error) { return key, nil })
		if err != nil || token.Valid != true {
			// 过期或者非正确处理
			log.Errorf(token)
			log.Error("令牌错误:"+err.Error())

			if protocol.ErrTokenExpired.Is(err.(*jwt.ValidationError).Inner) {
				w.Failed(protocol.ErrTokenExpired)
			} else {
				w.Failed(protocol.ErrTokenValidate)
			}
			return
		}
		w.AddContext("token",token.Raw)
		if claim, ok := token.Claims.(JwtClaimWithUser); ok {
			w.AddContext("user",claim.GetUser())
		}
		next.ServeHTTP(w,r)
	}
}

func SignToken(creator JwtClaimCreator,user interface{},a lokas.IProcess,expire time.Duration,rsa bool)(string,error){
	var claim *jwt.Token
	var key interface{}
	if rsa {
		key=a.Config().Get("SigningKeyPrivate")
		claim=jwt.NewWithClaims(jwt.SigningMethodRS256,creator(user,expire))
	} else {
		key=a.Config().Get("SigningKey")
		claim=jwt.NewWithClaims(jwt.SigningMethodHS256,creator(user,expire))
	}

	token,err:=claim.SignedString(key)
	if err != nil {
		log.Error("生成令牌出错:"+err.Error())
		return "",err
	}
	return token,nil
}

func JwtSign(rsa bool,creator JwtClaimCreator) func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler){
	return func(w rox.ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
		next.ServeHTTP(w, r)
		user := w.GetContext("user")
		if user == nil {
			log.Error("用户不存在,令牌无效")
			w.Failed(protocol.ErrTokenValidate)
			return
		}
		token, err := SignToken(creator,user, a,TOKEN_EXPIRE_TIME,rsa)
		if err != nil {
			log.Error("生成令牌出错")
			w.Response(http.StatusInternalServerError,"服务器繁忙")
			return
		}
		refresh_token, err := SignToken(creator,user, a,REFRESH_TOKEN_EXPIRE_TIME,rsa)
		if err != nil {
			log.Error("生成令牌出错")
			w.Failed(protocol.ErrTokenValidate)
			return
		}
		w.AddContent("token", token)
		w.AddContent("refresh_token", refresh_token)
	}
}