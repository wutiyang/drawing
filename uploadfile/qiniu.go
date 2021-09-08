package uploadfile

import (
	//"encoding/base64"
	"encoding/hex"
	"fmt"

	//"io"
	"crypto/md5"
	"math/rand"
	"time"

	"github.com/wutiyang/drawing/libs"

	//	"github.com/qiniu/api.v7/kodo"
	"golang.org/x/net/context"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

var (
	ACCESS_KEY    string
	SECRET_KEY    string
	BUCKET        string
	QINIUHOST     string
	WXMEDIAGETURL string
)

type confArr struct {
	conf []*libs.S1
}

var (
	ReloadConfTime time.Duration = 60
	//	log            *libs.Loger
	c      *libs.S1
	r      *libs.RedisFunc
	lognew *libs.Loger
	retry  int8
)

func init() {

	libs.InitInfo(map[string]string{"transapp": "conf/httpconfig.toml"})
	c = libs.CONF["transapp"]
	lognew = libs.LOG["transapp"]
	//kodo.SetMac(c.Qiniu.ACCESS_KEY, c.Qiniu.SECRET_KEY)
}

func UpFileByPath(filepath string) string {

	key := GetGuid(filepath)
	putPolicy := storage.PutPolicy{
		Scope: c.Qiniu.BUCKET,
	}
	mac := qbox.NewMac(c.Qiniu.ACCESS_KEY, c.Qiniu.SECRET_KEY)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	// 可选配置
	putExtra := storage.PutExtra{}
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, filepath, &putExtra)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println(ret.Key)
	return c.Qiniu.QINIUHOST + ret.Key
}

//远程url同步到七牛
//access_token  微信token
//media_id      微信media_id
//return 七牛存储资源url
//func UpFileByUrl(access_token string, media_id string) (interface{}, error) {
//	url := fmt.Sprintf(WXMEDIAGETURL, access_token, media_id)
//	key := GetGuid(media_id)
//	zone := 0
//	c := kodo.New(zone, nil)
//	bucket := c.Bucket(c.Qiniu.BUCKET)
//	ctx := context.Background()
//	err := bucket.Fetch(ctx, key, url)
//	if err != nil {
//		return nil, err
//	}
//	return key, nil
//}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func GetGuid(num string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num1 := fmt.Sprintf("%d", r.Intn(10000))
	num2 := fmt.Sprintf("%d", r.Intn(10000))
	num3 := fmt.Sprintf("%d", time.Now().UnixNano())

	uuid := num3 + num1 + num2 + num
	//加密
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(uuid))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
