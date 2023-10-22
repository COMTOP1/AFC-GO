package objects

import "strings"

type (
	JWTToken struct {
		Jti, Kid, Aud, Sub, Email, UserAgent, SessionId string
		Iat, Nbf, Exp                                   int64
	}
)

func (j *JWTToken) SetFields(Jti, Kid, Aud, Sub, Email, UserAgent, SessionId string, Iat, Nbf, Exp int64) {
	j.Jti = Jti
	j.Kid = Kid
	j.Aud = Aud
	j.Sub = Sub
	j.Email = Email
	j.UserAgent = UserAgent
	j.SessionId = SessionId
	j.Iat = Iat
	j.Nbf = Nbf
	j.Exp = Exp
}

func (j *JWTToken) GetJti() string {
	return j.Jti
}

func (j *JWTToken) SetJti(Jti string) {
	j.Jti = Jti
}

func (j *JWTToken) GetKid() string {
	return j.Kid
}

func (j *JWTToken) SetKid(Kid string) {
	j.Kid = Kid
}

func (j *JWTToken) GetAud() string {
	return j.Aud
}

func (j *JWTToken) SetAud(Aud string) {
	j.Aud = Aud
}

func (j *JWTToken) GetSub() string {
	return j.Sub
}

func (j *JWTToken) SetSub(Sub string) {
	j.Sub = Sub
}

func (j *JWTToken) GetEmail() string {
	return j.Email
}

func (j *JWTToken) SetEmail(Email string) {
	j.Email = Email
}

func (j *JWTToken) GetUserAgent() string {
	return j.UserAgent
}

func (j *JWTToken) SetUserAgent(UserAgent string) {
	j.UserAgent = UserAgent
}

func (j *JWTToken) GetSessionId() string {
	return j.SessionId
}

func (j *JWTToken) SetSessionId(SessionId string) {
	j.SessionId = SessionId
}

func (j *JWTToken) GetIat() int64 {
	return j.Iat
}

func (j *JWTToken) SetIat(Iat int64) {
	j.Iat = Iat
}

func (j *JWTToken) GetNbf() int64 {
	return j.Nbf
}

func (j *JWTToken) SetNbf(Nbf int64) {
	j.Nbf = Nbf
}

func (j *JWTToken) GetExp() int64 {
	return j.Exp
}

func (j *JWTToken) SetExp(Exp int64) {
	j.Exp = Exp
}

func (j *JWTToken) Validate(UserAgent, SessionId string, Time int64) bool {
	if len(j.Jti) != 0 && len(j.Kid) != 0 && len(j.Aud) != 0 && len(j.Sub) != 0 && len(j.Email) != 0 && strings.EqualFold(j.UserAgent, UserAgent) && strings.EqualFold(j.SessionId, SessionId) && j.Iat <= Time && j.Nbf <= Time {
		return j.Exp > Time
	}
	return false
}
