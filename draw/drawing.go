package draw

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"math"
	"mime/multipart"

	"github.com/wutiyang/drawing/draw/circlemask"

	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	//	"strings"
	"time"

	"github.com/wutiyang/drawing/libs"
	//	"github.com/wutiyang/drawing/uploadfile"

	"image/color"
	"strings"

	"github.com/golang/freetype"
	"github.com/nfnt/resize"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/segmentio/ksuid"
	"golang.org/x/image/font"
)

type ResultWechat struct {
	Errcode    int64   `json:"errcode"`
	Errmsg     string  `json:"errmsg"`
	Type       float64 `json:"type"`
	Media_id   string  `json:"media_id"`
	Created_at int64   `json:"created_at"`
}

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

//url获取图片设置
type ImageLoad struct {
	Ext      string //扩展名
	Image    image.Image
	FileName string //文件名称
	FilePath string //文件路径
	FileSize int    //文件大小
}

//最终生成图片设置
type Pic struct {
	M       *image.RGBA
	Name    string
	Path    string
	Ext     string
	Delete  int
	Quality int
}

//json中获取所需图片及文字素材
type Material struct {
	Quality    string      `json:"quality"`
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Delete     string      `json:"delete"`
	Ext        string      `json:"ext"`
	Background *BackGround `json:"background"`
	Picture    *[]Picture  `json:"picture"`
	Text       *[]Text     `json:"text"`
}

//背景设置
type BackGround struct {
	Url    string `json:"url"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

//图片设置
type Picture struct {
	Url     string `json:"url"`
	Width   string `json:"width"`
	Height  string `json:"height"`
	X       string `json:"x"`
	Y       string `json:"y"`
	IsRound string `json:"is_round"` // 是否画圆
	IsShow  string `json:"is_show"`  // 是否显示
}

//文字设置
type Text struct {
	Text   string `json:"text"`
	Size   string `json:"size"`
	Color  string `json:"color"`
	X      string `json:"x"`
	Y      string `json:"y"`
	R      string `json:"r"`
	G      string `json:"g"`
	B      string `json:"b"`
	A      string `json:"a"`
	IsShow string `json:"is_show"` // 是否显示
}

//文字参数
var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "font/msyh.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
)

func init() {

	libs.InitInfo(map[string]string{"transapp": "conf/httpconfig.toml"})
	c = libs.CONF["transapp"]
	lognew = libs.LOG["transapp"]
}

//第一步 获取图片url资源
func (p *Pic) GetUrl(url string) *ImageLoad {
	ret, err := http.Get(url)
	if err != nil {
		log.Println(url)
		status := map[string]string{}
		status["status"] = "400"
		status["url"] = url
		panic(err)
	}
	defer ret.Body.Close()

	body := ret.Body
	data, _ := ioutil.ReadAll(body)

	//创建文件
	fileName := ksuid.New().String()
	file, err := os.Create(p.Path + fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	//写入文件
	size, err := file.Write(data)
	//打开文件
	file, err = os.Open(file.Name())
	if err != nil {
		log.Println(url)
		status := map[string]string{}
		status["status"] = "400"
		status["url"] = url
		panic(status)
	}
	defer file.Close()
	//获取扩展名
	ext, err := GetImgExt(file.Name())
	img, _ := DecodePic(file, ext)
	images := new(ImageLoad)
	images.FileSize = size
	images.FilePath = p.Path + fileName
	images.FileName = fileName
	images.Image = img
	images.Ext = ext
	return images
}

//第二步 图片缩放
func (p *Pic) Scaling(images *ImageLoad, width int, height int) string {
	m := resize.Resize(uint(width), uint(height), images.Image, resize.Lanczos3)
	file, err := os.Create(images.FilePath)
	if err != nil {
		log.Fatal(err)
	}
	p.EncodePic(file, m, images.Ext)
	defer file.Close()
	//图片缩放
	if err != nil {
		panic(err)
	}
	//重命名
	os.Rename(file.Name(), images.FilePath+images.Ext)
	return file.Name() + images.Ext
}

//第三部 背景图片画图
func (p *Pic) BackGround(url string, width int, height int) {
	images := p.GetUrl(url)
	filePath := p.Scaling(images, width, height)
	imgb, _ := os.Open(filePath)
	ext, _ := GetImgExt(imgb.Name())
	if p.Ext == "" {
		p.Ext = ext
	}
	img, _ := DecodePic(imgb, ext)
	defer os.Remove(filePath)
	defer imgb.Close()
	b := img.Bounds()
	m := image.NewRGBA(b)
	draw.Draw(m, b, img, image.ZP, draw.Src)
	p.M = m
}

//第四部 合成素材图片画图
func (p *Pic) DrawPicture(url string, width int, height int, x int, y int) {
	qrcode_url := p.GetUrl(url)
	filePath := p.Scaling(qrcode_url, width, height)
	imgb, _ := os.Open(filePath)
	ext, _ := GetImgExt(imgb.Name())
	watermark, _ := DecodePic(imgb, ext)
	defer os.Remove(filePath)
	defer imgb.Close()
	offset := image.Pt(x, y)
	draw.Draw(p.M, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)
}

func (p *Pic) CreateImage() string {
	filePath := p.Path + p.Name + p.Ext
	imgw, _ := os.Create(filePath)
	p.EncodePic(imgw, p.M, p.Ext)
	defer imgw.Close()
	if p.Delete == 1 {
		defer os.Remove(filePath)
	}
	return filePath
}

//第五步 文字内容画图
func (p *Pic) fontRender(text string, size float64, colour int, x int, y int, r uint8, g uint8, b uint8, a uint8) {
	flag.Parse()
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}
	fg := image.Black
	if colour == 2 {
		fg = image.White
	}
	if r > 0 || g > 0 || b > 0 || a > 0 {
		if a == 0 {
			a = 255
		}
		fg = image.NewUniform(color.RGBA{r, g, b, a})
	}
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(size)
	c.SetClip(p.M.Bounds())
	c.SetDst(p.M)
	c.SetSrc(fg)

	switch *hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}

	// Draw the text.
	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	_, err = c.DrawString(text, pt)
	if err != nil {
		log.Println(err)
		return
	}
	pt.Y += c.PointToFixed(size * 0)
}

//获取图片类型
func GetImgExt(file string) (ext string, err error) {
	var headerByte []byte
	headerByte = make([]byte, 8)
	fd, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer fd.Close()
	_, err = fd.Read(headerByte)
	if err != nil {
		return "", err
	}
	xStr := fmt.Sprintf("%x", headerByte)
	switch {
	case xStr == "89504e470d0a1a0a":
		ext = ".png"
	case xStr == "0000010001002020":
		ext = ".ico"
	case xStr == "0000020001002020":
		ext = ".cur"
	case xStr[:12] == "474946383961" || xStr[:12] == "474946383761":
		ext = ".gif"
	case xStr[:10] == "0000020000" || xStr[:10] == "0000100000":
		ext = ".tga"
	case xStr[:8] == "464f524d":
		ext = ".iff"
	case xStr[:8] == "52494646":
		ext = ".ani"
	case xStr[:4] == "4d4d" || xStr[:4] == "4949":
		ext = ".tiff"
	case xStr[:4] == "424d":
		ext = ".bmp"
	case xStr[:4] == "ffd8":
		ext = ".jpg"
	case xStr[:2] == "0a":
		ext = ".pcx"
	default:
		ext = ""
	}
	return ext, nil
}

//图片写入
func (p *Pic) EncodePic(w io.Writer, m image.Image, ext string) {
	if p.Quality == 0 {
		p.Quality = 75
	}
	switch {
	case ext == ".jpg":
		jpeg.Encode(w, m, &jpeg.Options{p.Quality})

	case ext == ".png":
		png.Encode(w, m)
	case ext == ".gif":
		gif.Encode(w, m, &gif.Options{})
	default:
		jpeg.Encode(w, m, &jpeg.Options{p.Quality})
	}
}

//打开图片
func DecodePic(r io.Reader, ext string) (image.Image, error) {
	switch {
	case ext == ".jpg":
		return jpeg.Decode(r)
	case ext == ".png":
		return png.Decode(r)
	case ext == ".gif":
		return gif.Decode(r)
	default:
		return jpeg.Decode(r)
	}
}

//合成画图
func CreateImage(mate *Material) string {
	picture := new(Pic)
	//清晰度设置
	if mate.Quality != "" {
		picture.Quality = TypeInt(mate.Quality)
	} else {
		picture.Quality = 65
	}
	//文件路径设置
	if mate.Path != "" {
		picture.Path = mate.Path
	} else {
		picture.Path = c.PicPath
	}
	//文件名称设置
	if mate.Name != "" {
		picture.Name = mate.Name
	} else {
		picture.Name = ksuid.New().String()
	}
	//文件扩展名
	if mate.Ext != "" {
		comma := strings.Index(mate.Ext, ".")
		picture.Ext = mate.Ext
		if comma < 0 {
			picture.Ext = "." + mate.Ext
		}
	} else {
		//默认ext为背景图 ext
		picture.Ext = ""
	}
	//是否删除原图片
	if mate.Delete == "1" {
		picture.Delete = 1
	}

	picture.BackGround(mate.Background.Url, TypeInt(mate.Background.Width), TypeInt(mate.Background.Height))
	for _, pic := range *mate.Picture {
		picture.DrawPicture(pic.Url, TypeInt(pic.Width), TypeInt(pic.Height), TypeInt(pic.X), TypeInt(pic.Y))
	}
	for _, text := range *mate.Text {
		picture.fontRender(text.Text, TypeFloat64(text.Size), TypeInt(text.Color), TypeInt(text.X), TypeInt(text.Y), TypeUint8(text.R), TypeUint8(text.G), TypeUint8(text.B), TypeUint8(text.A))
	}

	return picture.CreateImage()
}

//字符串转int
func TypeInt(e interface{}) int {
	switch v := e.(type) {
	case int:
		return v
	case string:
		b, _ := strconv.Atoi(v)
		return b
	}
	return 0
}

//多类型转float64
func TypeFloat64(e interface{}) float64 {
	switch v := e.(type) {
	case int:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	case float64:
		return v
	}
	return 0
}

func TypeUint8(str string) uint8 {
	unit := TypeInt(str)
	return uint8(unit)
}

//获取json字符串 开始画图
func GetMaterial(ob []byte) string {
	t1 := time.Now().Unix()
	s := new(Material)
	ffjson.Unmarshal(ob, &s)
	filePath := CreateImage(s)
	fmt.Println("usetime", time.Now().Unix()-t1, "s")
	return filePath
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func UniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

func Run(appid string, ob []byte) (string, int64) {
	filePath := GetMaterial(ob)
	fmt.Println(filePath)

	//url := uploadfile.UpFileByPath(filePath)

	media_id, create_at := uploadwx(appid, filePath)
	//os.Exit(0)
	return media_id, create_at
}

func uploadwx(appid, filename string) (string, int64) {

	var media_id string
	var create_at int64
	media_id = ""
	create_at = 0
	//http获取
	accesstoken := libs.GetTokenByHttp(appid)

	if len(accesstoken) > 0 {
		Url := fmt.Sprintf(c.Url.WxUrl, accesstoken)
		// start send server.
		//work(task, dataType)
		fmt.Println(Url)
		media_id, create_at = Toupload(Url, filename)
	} else {
		lognew.Info("token", appid+" accesstoken is empty!")
	}
	return media_id, create_at
}

func Toupload(url, filename string) (string, int64) {

	var reswx ResultWechat
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, _ := bodyWriter.CreateFormFile("media", filename)

	file, _ := os.Open(filename)
	defer file.Close()

	io.Copy(fileWriter, file)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, _ := http.Post(url, contentType, bodyBuffer)
	defer resp.Body.Close()

	resp_body, _ := ioutil.ReadAll(resp.Body)

	if err := ffjson.Unmarshal([]byte(resp_body), &reswx); err != nil {
		lognew.Error("wechat", fmt.Sprintf(" result info error:", err))
	}

	log.Println(resp.Status)
	log.Println(string(resp_body))
	return reswx.Media_id, reswx.Created_at
}

//func Toupload(url, filename string) string {
//	var (
//		err error = nil
//		msg string
//	)

//	task := make(map[string]interface{})
//	var reswx ResultWechat

//	task["file"] = filename

//	h := &libs.HttpRequest{NetCount: c.Httpinfo.NetCount,
//		Ctime: c.Httpinfo.Ctime, Rwtime: c.Httpinfo.Rwtime}

//	msg, err = h.HttpPost(url, task, "postfile")

//	fmt.Println("==============", msg, err)

//	if err = ffjson.Unmarshal([]byte(msg), &reswx); err != nil {
//		lognew.Error("wechat", fmt.Sprintf(" result info error:", err))
//	}

//	//	if err != nil {
//	//		lognew.Error("wechat", fmt.Sprintf("send template error:%s %s", err, msg))
//	//	}
//	return reswx.Media_id
//}

func ArticleRun(appid string, ob []byte) (string, int64) {
	t1 := time.Now().Unix()
	s := new(Material)
	ffjson.Unmarshal(ob, &s)

	filePath := CreateImageInArticle(s)
	fmt.Println("usetime", time.Now().Unix()-t1, "s")
	fmt.Println(filePath)

	media_id, create_at := uploadwx(appid, filePath)
	//os.Exit(0)
	return media_id, create_at
}

//画圆-合成画图
func CreateImageInArticle(mate *Material) string {
	picture := new(Pic)
	//清晰度设置
	if mate.Quality != "" {
		picture.Quality = TypeInt(mate.Quality)
	} else {
		picture.Quality = 65
	}
	//文件路径设置
	if mate.Path != "" {
		fmt.Println("path is:", mate.Path)
		picture.Path = mate.Path
	} else {
		fmt.Println("path is default:", c.PicPath)
		picture.Path = c.PicPath
	}
	//文件名称设置
	if mate.Name != "" {
		picture.Name = mate.Name
	} else {
		picture.Name = ksuid.New().String()
	}
	//文件扩展名
	if mate.Ext != "" {
		comma := strings.Index(mate.Ext, ".")
		picture.Ext = mate.Ext
		if comma < 0 {
			picture.Ext = "." + mate.Ext
		}
	} else {
		//默认ext为背景图 ext
		picture.Ext = ""
	}
	//是否删除原图片
	if mate.Delete == "1" {
		picture.Delete = 1
	}

	picture.BackGround(mate.Background.Url, TypeInt(mate.Background.Width), TypeInt(mate.Background.Height))
	for _, pic := range *mate.Picture {
		if strings.ToUpper(pic.IsShow) != "Y" {
			continue
		}
		if strings.ToUpper(pic.IsRound) == "Y" { // 画圆处理
			picture.DrawCirclePicture(pic.Url, TypeInt(pic.Width), TypeInt(pic.Height), TypeInt(pic.X), TypeInt(pic.Y))
		} else {
			picture.DrawPicture(pic.Url, TypeInt(pic.Width), TypeInt(pic.Height), TypeInt(pic.X), TypeInt(pic.Y))
		}
	}
	for _, text := range *mate.Text {
		if strings.ToUpper(text.IsShow) != "Y" {
			continue
		}
		fmt.Println("write text", text.Text)
		// #000000颜色值转换RGB值
		r, g, b := dealColor(text)
		//picture.fontRender(text.Text, TypeFloat64(text.Size), TypeInt(text.Color), TypeInt(text.X), TypeInt(text.Y), TypeUint8(text.R), TypeUint8(text.G), TypeUint8(text.B), TypeUint8(text.A))
		picture.fontRender(text.Text, TypeFloat64(text.Size), TypeInt(text.Color), TypeInt(text.X), TypeInt(text.Y), TypeUint8(r), TypeUint8(g), TypeUint8(b), TypeUint8(text.A))
	}

	return picture.CreateImage()
}

// 画圆合成
func (p *Pic) DrawCirclePicture(url string, width int, height int, x int, y int) {
	// 远程url图片信息
	qrcode_url := p.GetUrl(url)
	// 图片缩放
	filePath := p.Scaling(qrcode_url, width, height)
	//
	imgb, _ := os.Open(filePath)
	ext, _ := GetImgExt(imgb.Name())

	watermark, _ := DecodePic(imgb, ext)
	defer os.Remove(filePath)
	defer imgb.Close()
	offset := image.Pt(x, y)

	// 遮罩处理
	w := watermark.Bounds().Max.X - watermark.Bounds().Min.X
	h := watermark.Bounds().Max.Y - watermark.Bounds().Min.Y
	// 圆半径
	d := w
	if w > h {
		d = h
	}
	fmt.Println("半径:", d)

	//把头像转成Png,否则会有白底
	srcPng := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(srcPng, watermark.Bounds(), watermark, watermark.Bounds().Min, draw.Over)

	// 遮罩处理
	maskImg := circlemask.NewCircleMask(srcPng, image.Point{0, 0}, d)

	// 合成
	draw.DrawMask(p.M, watermark.Bounds().Add(offset), maskImg, image.ZP, maskImg, image.ZP, draw.Over)
	//draw.Draw(p.M, watermark.Bounds().Add(offset), maskImg, image.ZP, draw.Over)
}

// 画圆算法
func circleMask(d int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, d, d))

	for x := 0; x < d; x++ {
		for y := 0; y < d; y++ {
			dis := math.Sqrt(math.Pow(float64(x-d/2), 2) + math.Pow(float64(y-d/2), 2))
			if dis > float64(d)/2 {
				img.Set(x, y, color.RGBA{255, 255, 255, 0})
			} else {
				img.Set(x, y, color.RGBA{0, 0, 255, 255})
			}
		}
	}
	return img
}

// 颜色处理
func dealColor(text Text) (red, green, blue string) {
	color_str := strings.TrimLeft(text.Color, "#")
	color_str = strings.TrimPrefix(color_str, "0x") //过滤掉16进制前缀

	color64, err := strconv.ParseInt(color_str, 16, 32) //字串到数据整型
	if err != nil {
		panic(err)
	}
	color32 := int(color64) //类型强转

	r, g, b := hexToRGB(color32)
	return strconv.Itoa(r), strconv.Itoa(g), strconv.Itoa(b)
}

/**
 * 颜色代码转换为RGB
 * input int
 * output int red, green, blue
 **/
func hexToRGB(color int) (red, green, blue int) {
	red = color >> 16
	green = (color & 0x00FF00) >> 8
	blue = color & 0x0000FF
	return
}
