package common

import "time"

const (
	VerificationCodeCookie = "vfcode_session"

	WebVFCodeLimit        = "verification_code:sid:%s" // 验证码存储
	WebVFCodeUnlock       = "verification_code:unlock:sid:%s"
	WebVFCodeLimitExpire  = 300 * time.Second
	WebVFCodeUnlockExpire = 3600 * time.Second
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type ErrorResponse struct {
	ErrNo  int    `json:"code"`
	ErrMsg string `json:"msg"`
}

func (e ErrorResponse) Error() string {
	return e.ErrMsg
}

var ErrVerificationCode = ErrorResponse{
	ErrNo:  6010,
	ErrMsg: "verification code err",
}

var ErrVerificationCodeIncorrect = ErrorResponse{
	ErrNo:  6011,
	ErrMsg: "code is incorrect",
}

var ErrorVerificationCodeExpired = ErrorResponse{
	ErrNo:  6012,
	ErrMsg: "code is expired",
}

var ErrorParamInvalid = ErrorResponse{
	ErrNo:  4000,
	ErrMsg: "param err",
}
