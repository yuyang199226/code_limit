package service

import (
	"bytes"
	"code_limit/common"
	"code_limit/helpers"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strings"

	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

const (
	codeNum       = 4
	defaultWidth  = 240
	defaultHeight = 80
)

var fontData = MustAsset("fonts/DejaVuSans-Bold.ttf")

type CodeService struct {
}

func NewCodeService() *CodeService {
	return &CodeService{}
}

func (e *CodeService) GenCode(ctx *gin.Context, codeSession string) (string, error) {
	code := helpers.GenerateRandomString(codeNum)
	b := e.ImgText(defaultWidth, defaultHeight, code)
	base64Str := base64.StdEncoding.EncodeToString(b)
	cacheKey := fmt.Sprintf(common.WebVFCodeLimit, codeSession)
	if err := helpers.RedisClient.Set(cacheKey, code, common.WebVFCodeLimitExpire).Err(); err != nil {
		logrus.Errorf("redis set k=%s, val=%s failed,err=%v", cacheKey, code, err)
		return "", common.ErrVerificationCode
	}
	return "data:image/jpeg;base64," + base64Str, nil
}

func (e *CodeService) Verify(ctx *gin.Context, codeSession string, inputCode string) error {
	cacheKey := fmt.Sprintf(common.WebVFCodeLimit, codeSession)
	res, err := helpers.RedisClient.Get(cacheKey).Result()
	if err != nil {
		logrus.Errorf("redis get k=%s err: %v", cacheKey, err)
		return common.ErrVerificationCode
	}
	if res == "" {
		return common.ErrorVerificationCodeExpired
	}
	dbCode := string(res)
	if len(inputCode) != codeNum {
		logrus.Warnf("input=%s, db=%s,len=%d", inputCode, dbCode, len(inputCode))
		return common.ErrVerificationCodeIncorrect
	}
	if strings.ToLower(inputCode) != strings.ToLower(dbCode) {
		logrus.Warnf("input=%s, db=%s", inputCode, dbCode)
		return common.ErrVerificationCodeIncorrect
	}
	// 验证通过
	return e.Unlock(ctx, codeSession)

}

func (e *CodeService) Unlock(ctx *gin.Context, codeSession string) error {
	key := fmt.Sprintf(common.WebVFCodeUnlock, codeSession)
	clintIp := ctx.ClientIP()
	err := helpers.RedisClient.Set(key, fmt.Sprintf("ip:%s", clintIp), common.WebVFCodeUnlockExpire).Err()
	if err != nil {
		logrus.Errorf("redis set k=%s err: %v", key, err)
		return common.ErrVerificationCode
	}
	return nil
}

func (e *CodeService) ImgText(width, height int, text string) (b []byte) {
	textLen := len(text)
	dc := gg.NewContext(width, height)
	bgR, bgG, bgB, bgA := getRandColorRange(240, 255)
	dc.SetRGBA255(bgR, bgG, bgB, bgA)
	dc.Clear()
	// 干扰线
	for i := 0; i < 20; i++ {
		x1, y1 := getRandPos(width, height)
		x2, y2 := getRandPos(width, height)
		r, g, b, a := getRandColor(255)
		w := float64(rand.Intn(3) + 1)
		dc.SetRGBA255(r, g, b, a)
		dc.SetLineWidth(w)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}

	fontSize := float64(height/2) + 5
	face := loadFontFace(fontSize)
	dc.SetFontFace(face)

	for i := 0; i < len(text); i++ {
		r, g, b, _ := getRandColor(100)
		dc.SetRGBA255(r, g, b, 255)
		fontPosX := float64(width/textLen*i) + fontSize*0.6

		writeText(dc, text[i:i+1], float64(fontPosX), float64(height/2))
	}
	buffer := bytes.NewBuffer(nil)
	dc.EncodePNG(buffer)
	b = buffer.Bytes()
	return
}

func writeText(dc *gg.Context, text string, x, y float64) {
	xfload := 5 - rand.Float64()*10 + x
	yfload := 5 - rand.Float64()*10 + y

	radians := 40 - rand.Float64()*80
	dc.RotateAbout(gg.Radians(radians), x, y)
	dc.DrawStringAnchored(text, xfload, yfload, 0.2, 0.5)
	dc.RotateAbout(-1*gg.Radians(radians), x, y)
	dc.Stroke()
}

// 随机坐标
func getRandPos(width, height int) (x float64, y float64) {
	x = rand.Float64() * float64(width)
	y = rand.Float64() * float64(height)
	return x, y
}

// 随机颜色
func getRandColor(maxColor int) (r, g, b, a int) {
	r = int(uint8(rand.Intn(maxColor)))
	g = int(uint8(rand.Intn(maxColor)))
	b = int(uint8(rand.Intn(maxColor)))
	a = int(uint8(rand.Intn(255)))
	return r, g, b, a
}

// 随机颜色范围
func getRandColorRange(miniColor, maxColor int) (r, g, b, a int) {
	if miniColor > maxColor {
		miniColor = 0
		maxColor = 255
	}
	r = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	g = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	b = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	a = int(uint8(rand.Intn(maxColor-miniColor) + miniColor))
	return r, g, b, a
}

// // 加载字体
func loadFontFace(points float64) font.Face {
	// 这里是将字体TTF文件转换成了 byte 数据保存成了一个 go 文件 文件较大可以到附录下
	// 通过truetype.Parse可以将 byte 类型的数据转换成TTF字体类型
	f, err := truetype.Parse(fontData)

	if err != nil {
		panic(err)
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: points,
	})
	return face
}
