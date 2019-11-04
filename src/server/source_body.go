package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"flag"


)
	var(
		file_name  = flag.String("file_name", "", "name of file post")
	)
type dataOfPostImage struct{

	Name string `json:"name"`
	Height string `json:"height"`
	Width string `json:"width"`


}
type Data struct{
	Name string `json:"name"`
	Watermark bool `json:"watermark"`
	Detail DetailWaterMark `json:detail`
	Size []Size `json:"size"`
	OriginalHeight string `json:"originalheight"`
	OriginalWidth string `json:"originalwidth"`
	Type string  	`json:"type"`
	Top  int 	`json:"top"`
	Left int        `json:"left"`
	Rotate int        `json:"rotate"`
	AreaWidth int	`json:"areawidth"`
	AreaHeight int	`json:"areaheight"`
	Factor    int       `json:"factor"`
	Unique    bool   `json:"unique"`
	UserId   int	`json:"userId"`
	DistributionId   string	`json:"distributionId,omitempty"`
	HostName string `json:"hostname,omitempty"`

}
type Size struct {
	Height int `json:"height"`
	Width int `json:"width"`
	MetaName string `json:"metaname"`
}
type DetailWaterMark struct {
	Width       int `json:"textwidth"`
	DPI         int	`json:"dpi"`
	Margin      int `json:"margin"`
	Opacity     float32 `json:"opacity"`
	NoReplicate bool     `json:"noreplicate"`
	Text        string   `json:"text"`
	Font        string   `json:"font"`
	Background  Color    `json:"background"`
}
type Color struct {
	R uint8 `json:"r"`
	G uint8	`json:"g"`
	B uint8	`json:"b"`
}
const formFieldName = "file"
const maxMemory int64 = 1024 * 1024 * 64

const ImageSourceTypeBody ImageSourceType = "payload"

type BodyImageSource struct {
	Config *SourceConfig
}

func NewBodyImageSource(config *SourceConfig) ImageSource {
	return &BodyImageSource{config}
}

func (s *BodyImageSource) Matches(r *http.Request) bool {
	return r.Method == "POST" || r.Method == "PUT"
}

func (s *BodyImageSource) GetImage(r *http.Request) ([]byte, error) {
	if isFormBody(r) {
		return readFormBody(r)
	}
	return readRawBody(r)
}

func isFormBody(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/")
}

func readFormBody(r *http.Request) ([]byte, error) {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return nil, err
	}


	file,header, err := r.FormFile("file")

	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)

	if len(buf) == 0 {
		err = ErrEmptyBody
	}
	*file_name=header.Filename




	return buf, err
}

func formField(r *http.Request) string {
	if field := r.URL.Query().Get("field"); field != "" {
		return field
	}
	return formFieldName
}

func readRawBody(r *http.Request) ([]byte, error) {

	return ioutil.ReadAll(r.Body)

}

func init() {
	RegisterSource(ImageSourceTypeBody, NewBodyImageSource)
}
