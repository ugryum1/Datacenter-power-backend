package ds

import "github.com/golang-jwt/jwt/v4"

type DatacenterEngineerJWTClaims struct {
	jwt.StandardClaims
	EngineerID uint `json:"engineer_id"`
}
