package main

import "github.com/dgrijalva/jwt-go"

type JwtClaims struct {
	jwt.StandardClaims
	Session  string   `json:"sid,omitempty"`
	Channels []string `json:"channels,omitempty"`
}
