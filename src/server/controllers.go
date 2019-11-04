package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/h2non/bimg.v1"
	"gopkg.in/h2non/filetype.v0"

	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/Machiel/slugify"
	"github.com/rs/xid"

	"github.com/dgrijalva/jwt-go"
)

type Sizer interface {
	Size() int64
}

func indexController(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorReply(r, w, ErrNotFound, ServerOptions{})
		return
	}

	body, _ := json.Marshal(CurrentVersions)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func uploadFileController(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if file != nil {
		if file.(Sizer).Size() > (10 * MB_FILE) {
			ErrorReply(r, w, ErrEntityTooLarge, ServerOptions{})
			return
		}
	}

	if err != nil {
		ErrorReply(r, w, ErrUnsupportedMedia, ServerOptions{})
		return
	}
	//doc file
	buf, err := ioutil.ReadAll(file)

	if len(buf) == 0 {
		ErrorReply(r, w, ErrEmptyBody, ServerOptions{})
		return
	}
	//de dam bao ten file anh ko co ki tu dac biet
	name_of_file := slugify.Slugify(header.Filename)
	//tao ten cua file
	guid := xid.New()
	//de dam bao key anh la duy nhat nen dung unique id cua thu vien xid
	name_of_file = name_of_file + guid.String()

	//lay h hien tai
	t := time.Now()
	//set location cho no
	utc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Println("[App.Error] : can't be set localtion time of image")
		return
	}
	string_time := t.In(utc).Format(time_layout)
	//upload file len amazone
	url, err := upload_File_Amazone(buf, header.Header.Get("Content-Type"), name_of_file, aws_path_file)
	//add vao struct fileimfo ket qua tra ve
	typeImage := ""
	if len(strings.Split(header.Header.Get("Content-Type"), "/")) > 0 {
		typeImage = strings.Split(header.Header.Get("Content-Type"), "/")[1]
	}
	reponse := fileImfo{header.Header.Get("Content-Type"), string_time, name_of_file, url, typeImage}

	body, _ := json.Marshal(reponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func attachmentsController(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if file != nil {
		if file.(Sizer).Size() > (20 * MB_FILE) {
			ErrorReply(r, w, ErrEntityTooLarge, ServerOptions{})
			return
		}
	}

	if err != nil {
		ErrorReply(r, w, ErrUnsupportedMedia, ServerOptions{})
		return
	}

	//doc file
	buf, err := ioutil.ReadAll(file)

	if len(buf) == 0 {
		ErrorReply(r, w, ErrEmptyBody, ServerOptions{})
		return
	}
	//de dam bao ten file anh ko co ki tu dac biet
	name_of_file := slugify.Slugify(header.Filename)
	//tao ten cua file
	guid := xid.New()
	//de dam bao key anh la duy nhat nen dung unique id cua thu vien xid
	name_of_file = name_of_file + guid.String()
	userGroupId, err := getGroupId(w, r)
	if err != nil {
		ErrorReply(r, w, ErrRoleUserNotAllowed, ServerOptions{})
		return
	}

	//neu groupId khac admin thi se so sanh voi groupId cua filter
	new_name := ""
	if userGroupId != "admin" {
		//kiem tra groupId
		filterGroupId := r.URL.Query().Get("groupId")
		if filterGroupId == "" {
			ErrorReply(r, w, ErrMissingFilterGroupId, ServerOptions{})
			return
		}
		//neu userGroupId khac voi filter GroupId thi se ko dc
		if userGroupId != filterGroupId {
			ErrorReply(r, w, ErrFilterGroupId, ServerOptions{})
			return
		}

		new_name = userGroupId + "/" + name_of_file

	} else {
		new_name = "0" + "/" + name_of_file
	}
	//lay h hien tai
	t := time.Now()
	//set location cho no
	utc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Println("[App.Error] : can't be set localtion time of image")
		return
	}
	string_time := t.In(utc).Format(time_layout)
	//upload file len amazone
	url, err := upload_AttachMents_Amazone(buf, header.Header.Get("Content-Type"), new_name, aws_path_attachments)
	//add vao struct fileimfo ket qua tra ve

	typeImage := ""
	if len(strings.Split(header.Header.Get("Content-Type"), "/")) > 0 {
		typeImage = strings.Split(header.Header.Get("Content-Type"), "/")[1]
	}
	reponse := fileImfo{header.Header.Get("Content-Type"), string_time, name_of_file, url, typeImage}

	body, _ := json.Marshal(reponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func publicUploadFileController(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	timeExpires := r.PostFormValue("timeExpires")
	tExp, tExpErr := strconv.Atoi(timeExpires)
	if file != nil {
		if file.(Sizer).Size() > (5 * MB_FILE) {
			ErrorReply(r, w, ErrEntityTooLarge, ServerOptions{})
			return
		}
	}

	if err != nil {
		ErrorReply(r, w, ErrUnsupportedMedia, ServerOptions{})
		return
	}
	if timeExpires == "" {
		ErrorReply(r, w, ErrMissingParamTimeExpires, ServerOptions{})
		return
	}
	if tExpErr != nil {
		ErrorReply(r, w, ErrTimeExpiresIsInvalid, ServerOptions{})
		return
	}
	if tExp == 0 {
		ErrorReply(r, w, ErrTimeExpiresCannotBeZero, ServerOptions{})
		return
	}
	//doc file
	buf, err := ioutil.ReadAll(file)

	if len(buf) == 0 {
		ErrorReply(r, w, ErrEmptyBody, ServerOptions{})
		return
	}
	//de dam bao ten file anh ko co ki tu dac biet
	name_of_file := slugify.Slugify(header.Filename)
	//tao ten cua file
	guid := xid.New()
	//de dam bao key anh la duy nhat nen dung unique id cua thu vien xidd
	name_of_file = name_of_file + guid.String()
	// userGroupId, err := getGroupId(w, r)
	if err != nil {
		ErrorReply(r, w, ErrRoleUserNotAllowed, ServerOptions{})
		return
	}

	//neu groupId khac admin thi se so sanh voi groupId cua filter
	new_name := "un-auth" + "/" + name_of_file

	//lay h hien tai
	t := time.Now()
	//set location cho no
	utc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Println("[App.Error] : can't be set localtion time of image")
		return
	}
	string_time := t.In(utc).Format(time_layout)
	//upload file len amazone
	url, err := publicUploadFileAmazon(buf, header.Header.Get("Content-Type"), new_name, aws_path_public_upload_files, aws_bucket_public_upload_files, tExp)
	//add vao struct fileimfo ket qua tra ve
	typeImage := ""
	if len(strings.Split(header.Header.Get("Content-Type"), "/")) > 0 {
		typeImage = strings.Split(header.Header.Get("Content-Type"), "/")[1]
	}
	reponse := fileImfo{header.Header.Get("Content-Type"), string_time, name_of_file, url, typeImage}

	body, _ := json.Marshal(reponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func uploadByEmailController(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	email := r.PostFormValue("email")
	//kiem tra email
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	if file != nil {
		if file.(Sizer).Size() > (5 * MB_FILE) {
			ErrorReply(r, w, ErrEntityTooLarge, ServerOptions{})
			return
		}
	}

	if err != nil {
		ErrorReply(r, w, ErrUnsupportedMedia, ServerOptions{})
		return
	}
	if email == "" {
		ErrorReply(r, w, ErrMissingParamEmail, ServerOptions{})
		return
	}
	if !Re.MatchString(email) {
		ErrorReply(r, w, ErrEmailAddressIsInvalid, ServerOptions{})
		return
	}
	//doc file
	buf, err := ioutil.ReadAll(file)

	if len(buf) == 0 {
		ErrorReply(r, w, ErrEmptyBody, ServerOptions{})
		return
	}
	//de dam bao ten file anh ko co ki tu dac biet
	name_of_file := slugify.Slugify(header.Filename)
	//tao ten cua file
	guid := xid.New()
	//de dam bao key anh la duy nhat nen dung unique id cua thu vien xid
	name_of_file = name_of_file + guid.String()

	//neu groupId khac admin thi se so sanh voi groupId cua filter

	//lay h hien tai
	t := time.Now()
	//set location cho no
	utc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Println("[App.Error] : can't be set localtion time of image")
		return
	}
	string_time := t.In(utc).Format(time_layout)
	//upload file len amazone
	url, err := uploadFileByEmailAmazon(buf, header.Header.Get("Content-Type"), name_of_file, aws_path_upload_email, aws_bucket_public_upload_files)
	//add vao struct fileimfo ket qua tra ve
	typeImage := ""
	if len(strings.Split(header.Header.Get("Content-Type"), "/")) > 0 {
		typeImage = strings.Split(header.Header.Get("Content-Type"), "/")[1]
	}
	reponse := fileImfo{header.Header.Get("Content-Type"), string_time, name_of_file, url, typeImage}

	body, _ := json.Marshal(reponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
func healthController(w http.ResponseWriter, r *http.Request) {
	health := GetHealthStats()
	body, _ := json.Marshal(health)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func imageController(o ServerOptions, operation Operation) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var imageSource = MatchSource(req)
		if imageSource == nil {
			ErrorReply(req, w, ErrMissingImageSource, o)
			return
		}

		buf, err := imageSource.GetImage(req)
		if err != nil {
			ErrorReply(req, w, NewError(err.Error(), BadRequest), o)
			return
		}

		if len(buf) == 0 {
			ErrorReply(req, w, ErrEmptyBody, o)
			return
		}

		imageHandler(w, req, buf, operation, o)
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request, buf []byte, Operation Operation, o ServerOptions) {

	//lay thong tin chieu dai va chieu rong cua anh goc
	info, err := getImageImfo(buf)
	s := r.FormValue("data")

	var data Data
	data.DistributionId = o.CFDistributionId
	data.Watermark = true
	json.Unmarshal([]byte(s), &data)
	//gank chieu rong va chieu cao vao data
	data.OriginalHeight = strconv.Itoa(info.Height)
	data.OriginalWidth = strconv.Itoa(info.Width)
	//chep phep thuoc tinh anh unique
	data.Unique = true
	//lay host name cua client request
	data.HostName = r.Header.Get("Origin")
	if data.HostName == "" {
		data.HostName = r.Header.Get("Referer")
	}
	log.Println(data.HostName)
	//neu khong kiem tra token thi se lay id  cua userId token vao data
	if o.ApiKey == "" && r.Method != "OPTIONS" {
		//kiem tra token
		jwt_token, err := newToken(w, r)
		if err == nil && jwt_token.Valid {
			// convert type claims thanh mot mang chua json cua token
			tokendata := jwt_token.Claims.(jwt.MapClaims)

			if tokendata["roles"] == nil {
				ErrorReply(r, w, ErrInvalidApiKey, o)
				return
			} else {
				//vi data la kieu interface {} khong sua dung duoc voi for nen ta se encode
				roles, err := json.Marshal(tokendata)

				if err != nil {
					log.Printf("[App.Error: can't marshal data in authorizeClient %s ", err)
					ErrorReply(r, w, Internal_Server_Error, o)
					return
				}
				//tiep theo tao mot interface roles moi va decoder du lieu vao
				var checkroles checkRoles
				err = json.Unmarshal(roles, &checkroles)

				if err != nil {
					log.Printf("[App.Error: can't marshal data in authorizeClient %s ", err)
					ErrorReply(r, w, Internal_Server_Error, o)
					return
				} else {
					data.UserId = checkroles.UserId
				}

			}
		} else {
			ErrorReply(r, w, ErrInvalidApiKey, o)
			return
		}
	}
	// Infer the body MIME type via mimesniff algorithm
	mimeType := http.DetectContentType(buf)

	// If cannot infer the type, infer it via magic numbers
	if mimeType == "application/octet-stream" {
		kind, err := filetype.Get(buf)
		if err == nil && kind.MIME.Value != "" {
			mimeType = kind.MIME.Value
		}
	}

	// Infer text/plain responses as potential SVG image
	if strings.Contains(mimeType, "text/plain") && len(buf) > 8 {
		if bimg.IsSVGImage(buf) {
			mimeType = "image/svg+xml"
		}
	}

	// Finally check if image MIME type is supported
	if IsImageMimeTypeSupported(mimeType) == false {
		ErrorReply(r, w, ErrUnsupportedMedia, o)
		return
	}

	opts := readParams(r.URL.Query(), &data)
	if opts.Type != "" && ImageType(opts.Type) == 0 {
		ErrorReply(r, w, ErrOutputFormat, o)
		return
	}
	//test nha
	image, err := Operation.Run(buf, opts, data)
	if err != nil {
		ErrorReply(r, w, NewError("Error while processing the image: "+err.Error(), BadRequest), o)
		return
	}

	w.Header().Set("Content-Type", image.Mime)
	w.Write(image.Body)
}

func formController(w http.ResponseWriter, r *http.Request) {
	operations := []struct {
		name   string
		method string
		args   string
	}{
		{"Resize", "resize", "width=300&height=200&type=jpeg"},
		{"Force resize", "resize", "width=300&height=200&force=true"},
		{"Crop", "crop", "width=300&quality=95"},
		{"SmartCrop", "crop", "width=300&height=260&quality=95&gravity=smart"},
		{"Extract", "extract", "top=100&left=100&areawidth=300&areaheight=150"},
		{"Enlarge", "enlarge", "width=1440&height=900&quality=95"},
		{"Rotate", "rotate", "rotate=180"},
		{"Flip", "flip", ""},
		{"Flop", "flop", ""},
		{"Thumbnail", "thumbnail", "width=100"},
		{"Zoom", "zoom", "factor=2&areawidth=300&top=80&left=80"},
		{"Color space (black&white)", "resize", "width=400&height=300&colorspace=bw"},
		{"Add watermark", "watermark", "textwidth=100&text=Hello&font=sans%2012&opacity=0.5&color=255,200,50"},
		{"Convert format", "convert", "type=png"},
		{"Image metadata", "info", ""},
	}

	html := "<html><body>"

	for _, form := range operations {
		html += fmt.Sprintf(`
    <h1>%s</h1>
    <form method="POST" action="/%s?%s" enctype="multipart/form-data">
      <input type="file" name="file" />
      <input type="submit" value="Upload" />
    </form>`, form.name, form.method, form.args)
	}

	html += "</body></html>"

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
