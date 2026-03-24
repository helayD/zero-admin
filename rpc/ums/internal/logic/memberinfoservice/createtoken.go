package memberinfoservicelogic

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 创建token
func createJwtToken(secretKey, name, mobile string, seconds, memberId int64) (string, error) {
	iat := time.Now().Unix()
	claims := jwt.MapClaims{
		"exp":        iat + seconds,
		"iat":        iat,
		"memberId":   memberId,
		"memberName": name,
		"mobile":     mobile,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
