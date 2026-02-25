package utils

import "github.com/lazygo/pkg/token/jwt"

var (
	ImageFormat = []string{".jpg", ".jpeg", ".png", ".bmp", ".gif"}
)

const PrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQC99rPobGTs6Nt4i2gAinzo0bN86c46Q0os5mY/nhnyswNV1JvW
v+IQT01o4NGiRlJUVvLbqeT7/J7UWyRvX4eoxlJGMiebNb+Se+ZFm5CHtL97QGNi
pMNyB1JnFiYVQC8lMMhPzDh/xWqiPTW4mUD0OWzhfVkGXdbN82NylcTMIQIDAQAB
AoGAC1nObjjr3uwbERzjXgToadD99JzT4c9bg0tRGMQNsN7ZeCt4FGDq1San7Xhh
jly9VqTLZizErfnuU6oOh0kiBe1WRjf+Yp6o3Lm/iYEDls3KWQ2MEoYtcmvv1cPy
YHAwZXQMUuAitg2g3PFcv/s2xbgzMsKIdqX9ECQlXN58EEECQQDWNRW9WKPKVQiq
7ZdhhcMSav6TCUrHG4XOIHaC6JXyJUgJbYod+2CKZOKKkpDAaCzdhnC4ijVcEiMd
C1fnyzUFAkEA4wa65f6mBwmQtLaSeP1xprM49TA//Qlreq6PUIz3KWAMdGDbWrgk
LTv4jrPNoom/UedLADeY4zFO+YwTGzelbQJBAME8mET8np1bQnt35DU8xvJphQg9
vLCLepStoluL2CDeUvL2Vz+e0sNmKRubBmwcTkT1D+DaBTLuxbBg2EcpgMUCQQDA
ooBKEcZdKE+iF1y3zw31jhOhcNyK38hUI/Q1PDeo4vgOc/UMyDziKQXiSE0PQuSV
IbXxLDvNY5SIiMPZj2ENAkBp7xNN/8xzi5Da7O3A9rK/DKwR6z5itJWedWlQlmZ5
auXhIcv7hbwQN/h5CsLfRLtUNvln6fUt3fGwMDURsZ1F
-----END RSA PRIVATE KEY-----
`

const PubKey = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC99rPobGTs6Nt4i2gAinzo0bN8
6c46Q0os5mY/nhnyswNV1JvWv+IQT01o4NGiRlJUVvLbqeT7/J7UWyRvX4eoxlJG
MiebNb+Se+ZFm5CHtL97QGNipMNyB1JnFiYVQC8lMMhPzDh/xWqiPTW4mUD0OWzh
fVkGXdbN82NylcTMIQIDAQAB
-----END PUBLIC KEY-----
`

var JwtDec, _ = jwt.NewJwtDecoder([]byte(PubKey))
