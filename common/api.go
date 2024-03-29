package common

type CheckVerificationCodeReq struct {
	Code string `json:"code"`
}

type GetVerificationCodeResp struct {
	ImgBase64 string `json:"img_base64"`
}
