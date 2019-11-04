package main

import (
	"encoding/json"
	"errors"

	"gopkg.in/h2non/bimg.v1"

	"log"
	"time"

	"path"

	"github.com/Machiel/slugify"
	"github.com/rs/xid"

	"strconv"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"

	//"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// Image stores an image binary buffer and its MIME type
type Image struct {
	Body []byte
	Mime string
}

// Operation implements an image transformation runnable interface
type Operation func([]byte, ImageOptions, Data) (Image, error)

// Run performs the image transformation
func (o Operation) Run(buf []byte, opts ImageOptions, data Data) (Image, error) {
	return o(buf, opts, data)
}

// ImageInfo represents an image details and additional metadata
type ImageInfo struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Type        string `json:"type"`
	Space       string `json:"space"`
	Alpha       bool   `json:"hasAlpha"`
	Profile     bool   `json:"hasProfile"`
	Channels    int    `json:"channels"`
	Orientation int    `json:"orientation"`
}

// dung de chua du lieu tra ve khi ma post thanh cong
type ImageImfomation struct {
	OriginalUrl    string       `json:"originalurl"`
	Title          string       `json:"title"`
	Watermark      bool         `json:"watermark"`
	Originalheight int          `json:"originalheight"`
	OriginalWidth  int          `json:"originalwidth"`
	OriginalType   string       `json:"originaltype"`
	AwsPath        string       `json:"awspath"`
	Height         int          `json:"height"`
	Width          int          `json:"width"`
	CndUrl         string       `json:"cndUrl"`
	ChildImage     []ChildImage `json:"childofimage"`
}
type ChildImage struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	FileType  string `json:"filetype"`
	UpdatedAt string `json:"updatedAt"`
	Title     string `json:"title"`
	Url       string `json:"url"`
	MetaName  string `json:"metaname"`
}

func Info(buf []byte, o ImageOptions, data Data) (Image, error) {
	// We're not handling an image here, but we reused the struct.
	// An interface will be definitively better here.
	image := Image{Mime: "application/json"}

	meta, err := bimg.Metadata(buf)
	if err != nil {
		return image, NewError("Cannot retrieve image metadata: %s"+err.Error(), BadRequest)
	}

	info := ImageInfo{
		Width:       meta.Size.Width,
		Height:      meta.Size.Height,
		Type:        meta.Type,
		Space:       meta.Space,
		Alpha:       meta.Alpha,
		Profile:     meta.Profile,
		Channels:    meta.Channels,
		Orientation: meta.Orientation,
	}

	body, _ := json.Marshal(info)
	image.Body = body

	return image, nil
}

func Handle_Image(buf []byte, opts bimg.Options, data Data) (Image, error) {

	//tao info de chua thong tin tra ve cho nguoi dung duoi dang object json
	var err error
	var info ImageImfomation
	//tao mot image voi du lieu tra ve la dang json
	image := Image{Mime: "application/json"}

	//gank anh goc
	original_opts := opts
	//neu co water mark thi anh goc se water mark
	if data.Watermark == true {
		original_opts, err = MakeOptionsWaterMarkImage(opts)
		if err != nil {
			log.Printf("[App.errors]: can't  watermark image in Resizes : %s", err)
			return image, err
		}
	} //sau do se duoc luu tren amzone
	original, err := ProcessWithArrayImage(buf, original_opts, data, "")
	if err != nil {
		log.Printf("[App.errors]: can't get origianl imfomation with ProcessWithArrayImage image in Resizes : %s", err)
		return image, err
	}
	//truyen du lieu anh goc vao info
	info.Title = original.Title
	info.OriginalUrl = original.Url
	info.Watermark = data.Watermark
	info.Originalheight = opts.Height
	info.OriginalWidth = opts.Width
	info.OriginalType = original.FileType
	info.AwsPath = aws_path_img
	info.CndUrl = original.Url
	info.Height = original.Height
	info.Width = original.Width
	//xu ly voi array height va width
	for _, s := range data.Size {

		opts.Width = s.Width
		opts.Height = s.Height
		new_options := opts
		//tuong tu neu co water mark thi xu ly
		if data.Watermark == true {
			new_options, err = MakeOptionsWaterMarkImage(opts)
			if err != nil {
				log.Printf("[App.errors]: can't  watermark image in Resizes : %s", err)
				return Image{}, err
			}

		}
		//se tra ve object childimage la con cua thang info
		childimage, err := ProcessWithArrayImage(buf, new_options, data, s.MetaName)
		if err != nil {
			log.Printf("[App.errors]: can't get childimage imfomation with ProcessWithArrayImage image in Resizes : %s", err)
			return image, err
		}
		//info se cap nhat them childeimage
		info.ChildImage = append(info.ChildImage, childimage)

	}
	//chuyen ve byte
	body, _ := json.Marshal(info)
	image.Body = body
	return image, nil
}

func Resize(buf []byte, o ImageOptions, data Data) (Image, error) {

	if o.Width == 0 && o.Height == 0 {
		return Image{}, NewError("Missing required param: height or width", BadRequest)
	}

	opts := BimgOptions(o)
	opts.Embed = true
	opts.Force = true
	if o.NoCrop == false {
		opts.Crop = true
	}
	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Resize : %s", err)
		return image, err
	}
	return image, nil
}

func fromexternalUpload(buf []byte, o ImageOptions, data Data) (Image, error) {

	var image, errors = Resize(buf, o, data)
	return image, errors
}

func Enlarge(buf []byte, o ImageOptions, data Data) (Image, error) {
	if o.Width == 0 || o.Height == 0 {
		return Image{}, NewError("Missing required params: height, width", BadRequest)
	}

	opts := BimgOptions(o)
	opts.Enlarge = true

	if o.NoCrop == false {
		opts.Crop = true
	}
	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Enlarge : %s", err)
		return image, err
	}
	return image, nil
}

func Extract(buf []byte, o ImageOptions, data Data) (Image, error) {

	o.AreaWidth = data.AreaWidth
	o.AreaHeight = data.AreaHeight

	if o.AreaWidth == 0 || o.AreaHeight == 0 {
		return Image{}, NewError("Missing required params: areawidth or areaheight", BadRequest)
	}

	opts := BimgOptions(o)
	opts.Top = data.Top
	opts.Left = data.Left
	opts.AreaWidth = data.AreaWidth
	opts.AreaHeight = data.AreaHeight
	opts.Embed = true

	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Extract : %s", err)
		return image, err
	}
	return image, nil
}

func Crop(buf []byte, o ImageOptions, data Data) (Image, error) {

	if o.Width == 0 && o.Height == 0 {
		return Image{}, NewError("Missing required param: height or width", BadRequest)
	}

	opts := BimgOptions(o)
	opts.Crop = true

	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Crop : %s", err)
		return image, err
	}
	return image, nil
}
func Profile(buf []byte, o ImageOptions, data Data) (Image, error) {

	if o.Width == 0 && o.Height == 0 {
		return Image{}, NewError("Missing required param: height or width", BadRequest)
	}

	opts := BimgOptions(o)
	opts.Crop = true
	data.Unique = false

	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Crop : %s", err)
		return image, err
	}
	// Get back infor from uploaded body
	var uploadedObject ImageImfomation
	err = json.Unmarshal(image.Body, &uploadedObject)
	if err == nil {
		// Invalidate on cloud front
		invalidateErr := invalidateProfileOnCloudFront(uploadedObject.Title, data.DistributionId)
		if invalidateErr != nil {
			log.Printf("[App.warnings]: Cannnot clear profile cache: %s", err)
		}
	}
	return image, nil
}

func invalidateProfileOnCloudFront(path string, distributionId string) error {
	// Init credntial infor
	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_token)
	_, err := creds.Get()
	if err != nil {
		log.Printf("[App.Errors] : bad credentials: %s", err)
		return err
	}
	cfg := aws.NewConfig().WithRegion(aws_region).WithCredentials(creds)
	// Init cloud front service
	svc := cloudfront.New(session.New(), cfg)
	// File path
	filePath := "/" + path
	// Init invalidation input
	timeStampString := time.Now().Format("20060102150405")
	invalidationQuantity := int64(1)
	input := &cloudfront.CreateInvalidationInput{}
	input.DistributionId = &distributionId
	input.InvalidationBatch = &cloudfront.InvalidationBatch{
		CallerReference: &timeStampString,
		Paths: &cloudfront.Paths{
			Quantity: &invalidationQuantity,
			Items:    []*string{&filePath},
		},
	}
	// Create a invalidation
	_, err = svc.CreateInvalidation(input)

	if err != nil {
		log.Printf("[App.Errors] : Error during creating invalidation: %v", err)
		log.Printf("[App.Errors] : Error during invalid: %s", path)
		if aerr, ok := err.(awserr.Error); ok {
			return aerr
		}
		return err
	}
	return nil
}

func Rotate(buf []byte, o ImageOptions, data Data) (Image, error) {
	o.Rotate = data.Rotate
	if o.Rotate == 0 {
		return Image{}, NewError("Missing required param: rotate", BadRequest)
	}

	opts := BimgOptions(o)
	opts.Embed = true

	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Rotate : %s", err)
		return image, err
	}
	return image, nil
}

func Flip(buf []byte, o ImageOptions, data Data) (Image, error) {
	opts := BimgOptions(o)
	opts.Flip = true
	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Flip : %s", err)
		return image, err
	}
	return image, nil
}

func Flop(buf []byte, o ImageOptions, data Data) (Image, error) {
	opts := BimgOptions(o)
	opts.Flop = true
	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Flop : %s", err)
		return image, err
	}
	return image, nil
}

func Thumbnail(buf []byte, o ImageOptions, data Data) (Image, error) {
	if o.Width == 0 && o.Height == 0 {
		return Image{}, NewError("Missing required params: width or height", BadRequest)
	}

	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, BimgOptions(o), data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Thumbnail : %s", err)
		return image, err
	}
	return image, nil
}

func Zoom(buf []byte, o ImageOptions, data Data) (Image, error) {

	o.Factor = data.Factor
	o.Top = data.Top
	o.Left = data.Left
	o.AreaHeight = data.AreaHeight
	o.AreaWidth = data.AreaWidth
	if o.Factor == 0 {
		return Image{}, NewError("Missing required param: factor", BadRequest)
	}
	opts := BimgOptions(o)

	if o.Top > 0 || o.Left > 0 {
		if o.AreaWidth == 0 && o.AreaHeight == 0 {
			return Image{}, NewError("Missing required params: areawidth, areaheight", BadRequest)
		}

		opts.Top = o.Top
		opts.Left = o.Left
		opts.AreaWidth = o.AreaWidth
		opts.AreaHeight = o.AreaHeight

		if o.NoCrop == false {
			opts.Crop = true
		}
	}

	opts.Zoom = o.Factor
	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Zoom : %s", err)
		return image, err
	}
	return image, nil
}

func Convert(buf []byte, o ImageOptions, data Data) (Image, error) {
	o.Type = data.Type
	if o.Type == "" {
		return Image{}, NewError("Missing required param: type", BadRequest)
	}
	if ImageType(o.Type) == bimg.UNKNOWN {
		return Image{}, NewError("Invalid image type: "+o.Type, BadRequest)
	}

	opts := BimgOptions(o)

	//ham xu ly chinh dung de xu ly cho tat ca cac api
	image, err := Handle_Image(buf, opts, data)
	if err != nil {
		log.Printf("[App.errors]: can't  Handle_Image in Convert : %s", err)
		return image, err
	}
	return image, nil
}

func Watermark(buf []byte, o ImageOptions, data Data) (Image, error) {

	if o.Text == "" {
		return Image{}, NewError("Missing required param: text", BadRequest)
	}

	opts := BimgOptions(o)
	opts.Watermark.DPI = o.DPI
	opts.Watermark.Text = o.Text
	opts.Watermark.Font = o.Font
	opts.Watermark.Margin = o.Margin
	opts.Watermark.Width = o.TextWidth
	opts.Watermark.Opacity = o.Opacity
	opts.Watermark.NoReplicate = o.NoReplicate

	if len(o.Color) > 2 {
		opts.Watermark.Background = bimg.Color{o.Color[0], o.Color[1], o.Color[2]}
	}

	return Process(buf, opts)
}

func Process(buf []byte, opts bimg.Options) (out Image, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch value := r.(type) {
			case error:
				err = value
			case string:
				err = errors.New(value)
			default:
				err = errors.New("libvips internal error")
			}
			out = Image{}
		}
	}()

	buf, err = bimg.Resize(buf, opts)
	if err != nil {
		return Image{}, err
	}

	mime := GetImageMimeType(bimg.DetermineImageType(buf))
	return Image{Body: buf, Mime: mime}, nil
}

func buildImfomation(opts bimg.Options, types string, url string, data Data, metaname string) (ChildImage, error) {
	childimage := ChildImage{}
	//lay h hien tai
	t := time.Now()
	//set location cho no
	utc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Println("[App.Error] : can't be set localtion time of image")
		return childimage, err
	}
	string_time := t.In(utc).Format(time_layout)
	info := ChildImage{
		Width:     opts.Width,
		Height:    opts.Height,
		FileType:  types,
		UpdatedAt: string_time,
		Title:     data.Name,
		Url:       url,
		MetaName:  metaname,
	}

	return info, nil

}
func getImageImfo(buf []byte) (ImageInfo, error) {
	meta, err := bimg.Metadata(buf)
	if err != nil {
		return ImageInfo{}, NewError("Cannot retrieve image metadata: %s"+err.Error(), BadRequest)
	}
	//lay du lieu cua anh orginal
	info := ImageInfo{
		Width:       meta.Size.Width,
		Height:      meta.Size.Height,
		Type:        meta.Type,
		Space:       meta.Space,
		Alpha:       meta.Alpha,
		Profile:     meta.Profile,
		Channels:    meta.Channels,
		Orientation: meta.Orientation,
	}
	return info, nil
}

func ProcessWithArrayImage(buf []byte, opts bimg.Options, data Data, metaname string) (out ChildImage, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch value := r.(type) {
			case error:
				err = value
			case string:
				err = errors.New(value)
			default:
				err = errors.New("libvips internal error")
			}
			out = ChildImage{}
		}
	}()

	buf, err = bimg.Resize(buf, opts)

	if err != nil {
		log.Printf("[App.errors]: can't bimg.Resize in ProcessWithArrayImage : %s", err)
		return ChildImage{}, err
	}
	//neu ten file tren vao rong thi se gank ten file bang ten anh
	if len(data.Name) == 0 {
		data.Name = path.Base(*file_name)
	}
	//kiem tra neu ten file van chua co
	if len(data.Name) == 0 {
		return ChildImage{}, errors.New("[App.Error]:Your Image name is NULL ")
	}
	//de dam bao ten file anh ko co ki tu dac biet
	data.Name = slugify.Slugify(data.Name)
	//lay mime cua image vi du nhu image/jpg
	mime := GetImageMimeType(bimg.DetermineImageType(buf))
	//lay types cua anh
	types := bimg.ImageTypes[bimg.DetermineImageType(buf)]
	//tao ten cua file
	guid := xid.New()
	name_of_file := "avatar" + strconv.Itoa(data.UserId)
	if data.Unique == true {
		//de dam bao key anh la duy nhat nen dung unique id cua thu vien xid
		name_of_file = data.Name + guid.String() + "." + types
	}

	url, err := upload_File_Amazone(buf, mime, name_of_file, aws_path_img)
	if err != nil {
		log.Printf("[App.errors]: can't upload_Image_Amazone in ProcessWithArrayImage : %s", err)
		return ChildImage{}, err
	}

	//gan lai ten cua file
	data.Name = name_of_file
	//tra ve thong tin cua image do va up len amazon
	childimage, err := buildImfomation(opts, types, url, data, metaname)

	if err != nil {
		log.Printf("[App.errors]: can't buildImfomation in ProcessWithArrayImage : %s", err)
		return childimage, err
	}
	return childimage, nil
}
func MakeOptionsWaterMarkImage(opts bimg.Options) (bimg.Options, error) {
	buffer, err := bimg.Read(image_watermark_path)
	if err != nil {
		log.Printf("[App.errors]: can't read watermark image in WaterMarkWithImage : %s", err)
		return opts, err
	}
	info_watermark, err := getImageImfo(buffer)
	if err != nil {
		log.Printf("[App.errors]: can't get imfo ofwatermark image in ` : %s", err)
		return opts, err
	}
	//xu ly khoang cach khi dat anh watermark vao
	opts.WatermarkImage.Left = opts.Width - (info_watermark.Width + 2)
	opts.WatermarkImage.Top = opts.Height - (info_watermark.Height + 2)
	opts.WatermarkImage.Buf = buffer
	opts.WatermarkImage.Opacity = 1.0

	if err != nil {
		log.Printf("[App.errors]: can't Handle Process in WaterMarkWithImage : %s", err)
		return opts, err
	}
	//tra ve options cua image do

	return opts, nil
}
