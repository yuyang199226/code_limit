package middleware

import (
	"code_limit/common"
	"code_limit/helpers"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RATELIMIT_UID  = "%s_RATELIMIT_UID_%d"
	RATELIMIT_IP   = "%s_RATELIMIT_IP_%s"
	RATELIMIT_CUID = "%s_RATELIMIT_CUID_%s"
)

const (
	FullIpMode    = 0
	SegmentIpMode = 3
)

type RateLimitRule struct {
	Key         string
	fillInteval time.Duration
	limitNum    int64
	level       int // 0验证码 1封禁
	ipMode      int //0, 3
}

var whiteUas = []string{"googlebot", "applebot", "bingbot", "duckDuckbot", "naverbot",
	"twitterbot", "yandex", "facebot",
	"facebookexternalhit", "facebookcatalog", "serpstatbot", "ahrefsbot", "ahrefssiteaudit"}

var blackIPs = []string{"10.12.12.12"}

var resourceRule = []RateLimitRule{
	//{"uid", 1 * time.Minute, 5}, // 测试demo
	//{"ip", 1 * time.Minute, 5},
	{"ip", 1 * time.Hour, 200, 0, 0},
	{"ip", 1 * time.Hour, 600, 0, 3},
	{"ip", 24 * time.Hour, 1000, 1, 0},
}

var RateLimitUrlConf = map[string][]RateLimitRule{
	"/questionai/chatweb/questionDetail": resourceRule,
	"/questionai/chatweb/getRelatedQuestions": []RateLimitRule{
		{"ip", 1 * time.Hour, 200, 0, 0},
		{"ip", 1 * time.Hour, 600, 0, 3},
		{"ip", 24 * time.Hour, 1000, 1, 0},
	},
	"/questionai/chatweb/questionList": []RateLimitRule{
		//{"uid", 1 * time.Minute, 5}, // 测试demo
		//{"ip", 1 * time.Minute, 5},
		{"ip", 1 * time.Hour, 200, 0, 0},
		{"ip", 1 * time.Hour, 600, 0, 3},
		{"ip", 24 * time.Hour, 1000, 1, 0},
	},
}

// type RateLimitUrlConf map[string]interface{}
func RateLimitCheck(ctx *gin.Context) {

	ua := strings.ToLower(ctx.GetHeader("User-Agent"))
	for _, seoUa := range whiteUas {
		if strings.Contains(ua, seoUa) {
			ctx.Next()
			return
		}
	}
	clintIp := ctx.ClientIP()
	//内部ip
	if strings.HasPrefix(clintIp, "10.") {
		ctx.Next()
		return
	}

	if helpers.IsStrIn(clintIp, blackIPs) {
		ctx.AbortWithStatusJSON(200, common.Response{1001, "disable"})
		return
	}
	// 如果验证码通过了
	if isUnlock(ctx, clintIp) {
		ctx.Next()
		return
	}

	flag := true

	//存在配置
	requestUrl := ctx.Request.URL.Path

	var checkKey string
	if _, ok := RateLimitUrlConf[requestUrl]; ok {
		cuid := "xxx"

		for _, v := range RateLimitUrlConf[requestUrl] {
			checkKey = ""
			//存在用户在检查
			if v.Key == "uid" {
				uid := 0
				if uid > 0 {
					checkKey = fmt.Sprintf(RATELIMIT_UID, requestUrl, uid)
				}
			}
			if v.Key == "ip" {
				//if strings.Contains(ctx.Request.Host, "mathai-svc.zhiji") {
				//	//clintIp = ctx.GetHeader("User-ClientIp")
				//}
				if v.ipMode == SegmentIpMode {
					ipSeg := getIPSeg(clintIp)
					if ipSeg == clintIp {
						logrus.Errorf("ip_segment failed =%s", clintIp)
					}
					checkKey = fmt.Sprintf(RATELIMIT_IP, requestUrl, ipSeg)
				} else {
					checkKey = fmt.Sprintf(RATELIMIT_IP, requestUrl, clintIp)
				}

				logrus.Infof("RateLimitCheck checkKey=%s", checkKey)
			}

			if v.Key == "cuid" {
				checkKey = fmt.Sprintf(RATELIMIT_CUID, requestUrl, cuid)
			}

			if len(checkKey) > 0 {
				flag = RateLimit(ctx, checkKey, v.fillInteval, v.limitNum)
				//触发5次报警,看是否加黑名单
				//zlog.Debugf(ctx, "[RateLimitCheck] %s, %s %+V", requestUrl, checkKey, flag)
				if !flag {
					errMsg := fmt.Sprintf("[RateLimitCheck] %s, %s %+v, ua=%s", requestUrl, checkKey, flag, ua)
					logrus.Warn(errMsg)
					fmt.Println("send alert")
					if v.level == 0 {
						// 验证码
						ctx.AbortWithStatusJSON(200, common.Response{1003, "too frequently"}) //1003 太频繁
					} else {
						// 禁止
						ctx.AbortWithStatusJSON(200, common.Response{1001, "disable"})
					}
					return
				}
			}
		}
	}
	ctx.Next()
}

// @param key string object for rate limit such as uid/ip+url
// @param fillInterval time.Duration such as 1*time.Second
// @param limitNum max int64 allowed number per fillInterval
// @return whether reach rate limit, false means reach
func RateLimit(ctx *gin.Context, key string, fillInterval time.Duration, limitNum int64) bool {
	//current tick time window
	tick := int64(time.Now().Unix() / int64(fillInterval.Seconds()))
	currentKey := fmt.Sprintf("mathai:RateLimit_%s_%d_%d_%d", key, fillInterval, limitNum, tick)
	startCount := 0
	_, err := helpers.RedisClient.SetNX(currentKey, startCount, fillInterval).Result()
	if err != nil {
		return true
		//panic(err)
	}
	//number in current time window
	quantum, err := helpers.RedisClient.Incr(currentKey).Result()
	//log.Println("结果", currentKey, quantum, limitNum)
	if err != nil {
		return true
		//panic(err)
	}
	if quantum > limitNum {
		return false
	}
	return true
}

func isUnlock(ctx *gin.Context, ip string) bool {
	codeSession, err := ctx.Cookie(common.VerificationCodeCookie)
	if err != nil {
		logrus.Warnf("get code vf_session failed ip=%s", ip)
		return false
	}
	if codeSession == "" {
		return false
	}
	key := fmt.Sprintf(common.WebVFCodeUnlock, codeSession)
	res, err := helpers.RedisClient.Get(key).Result()
	if err != nil {
		logrus.Errorf("redis get key=%s failed %v", key, err)
		return false
	}
	return len(res) > 0
}

// 192.168.19.0 -> 192.168.19
// ipv6 ?
func getIPSeg(ip string) string {
	ipSegs := strings.Split(ip, ".")
	if len(ipSegs) == 4 {
		ipSeg := strings.Join(ipSegs[:3], ".")
		return ipSeg
	}
	return ip
}
