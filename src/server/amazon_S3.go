package main

import (
	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"fmt"

	"bytes"

	"github.com/aws/aws-sdk-go/aws/awsutil"

	"net/url"
	"time"
)

type fileImfo struct {
	FileType  string `json:"filetype"`
	CreatedAt string `json:"createdAt"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Type      string `json:"type"`
}

//danh cho upload co duong duong host https::jinn.vn
func upload_File_Amazone(buf []byte, mime string, file_name string, aws_path string) (string, error) {

	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_token)
	_, err := creds.Get()
	if err != nil {
		fmt.Printf("[App.Error] : bad credentials: %s", err)
		return "", err
	}
	cfg := aws.NewConfig().WithRegion(aws_region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)
	//tao filebytes tu gia buf truyen vao
	fileBytes := bytes.NewReader(buf)
	//dung dan luu tru anh trong amazon s3
	path := "/" + aws_path + "/" + file_name
	cacheControl := "public, max-age=" + maxAGE
	params := &s3.PutObjectInput{
		ACL:          aws.String(s3.ObjectCannedACLPublicRead), //cho phep public ,xem duoc va download dc
		Bucket:       aws.String(aws_bucket),                   //no nam trong bucket gi
		Key:          aws.String(path),                         //voi key gi
		Body:         fileBytes,                                // phan byte cua anh truyen vao body
		CacheControl: aws.String(cacheControl),
		ContentType:  aws.String(mime), //loai anh
	}

	resp, err := svc.PutObject(params) // day putobjecinput len amazone
	if err != nil {
		fmt.Printf("[App.Error] : bad response: %s", err)
		return awsutil.StringValue(resp), err
	}

	//tao mot string url moi
	new_url := cdn_url + "/" + file_name

	return new_url, nil
}
func checkAndHandleFileSpecialMime(path string, mime string) string {
	newPath := path
	if mime == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		newPath = newPath + ".docx"
	} else if mime == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		newPath = newPath + ".xlsx"
	}
	return newPath
}

//ham tra ve voi duong dan truc tiep la host amazone
func upload_AttachMents_Amazone(buf []byte, mime string, file_name string, aws_path string) (string, error) {

	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_token)
	_, err := creds.Get()
	if err != nil {
		fmt.Printf("[App.Error] : bad credentials: %s", err)
		return "", err
	}
	cfg := aws.NewConfig().WithRegion(aws_region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)
	//tao filebytes tu gia buf truyen vao
	fileBytes := bytes.NewReader(buf)
	//dung dan luu tru anh trong amazon s3
	path := "/" + aws_path + "/" + file_name
	path = checkAndHandleFileSpecialMime(path, mime)
	cacheControl := "public, max-age=" + maxAGE

	params := &s3.PutObjectInput{
		ACL:          aws.String(s3.ObjectCannedACLPublicRead), //cho phep public ,xem duoc va download dc
		Bucket:       aws.String(aws_bucket_attachments),       //no nam trong bucket gi
		Key:          aws.String(path),                         //voi key gi
		Body:         fileBytes,                                // phan byte cua anh truyen vao body
		CacheControl: aws.String(cacheControl),
		ContentType:  aws.String(mime), //loai anh
	}
	resp, err := svc.PutObject(params) // day putobjecinput len amazone
	if err != nil {
		fmt.Printf("[App.Error] : bad response: %s", err)
		return awsutil.StringValue(resp), err
	}

	//lay link url cua anh tu amazone s3
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(aws_bucket_attachments), //nam o dau
		Key:    aws.String(path),                   //file ten gi
	})
	//lay chuoi url ra
	url_string, err := req.Presign(300 * time.Second)

	if err != nil {
		fmt.Printf("[App.Error] : can't get url from amazone s3: %s", err)
		return "", err
	}
	//do url co cac paramater tu amazone dinh kem nen doan code sau se loai bo cac parameter do
	url, err := url.Parse(url_string)
	if err != nil {
		fmt.Printf("[App.Error] : can't parse string to url: %s", err)
		return "", err
	}
	new_url := url.Host + url.Path
	return new_url, nil
}

func publicUploadFileAmazon(buf []byte, mime string, file_name string, aws_path string, awsBucket string, timeExpries int) (string, error) {

	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_token)
	_, err := creds.Get()
	if err != nil {
		fmt.Printf("[App.Error] : bad credentials: %s", err)
		return "", err
	}
	cfg := aws.NewConfig().WithRegion(aws_region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)
	//tao filebytes tu gia buf truyen vao
	fileBytes := bytes.NewReader(buf)
	//dung dan luu tru anh trong amazon s3
	path := "/" + aws_path + "/" + file_name
	cacheControl := "public, max-age=" + maxAGE
	params := &s3.PutObjectInput{
		ACL:          aws.String(s3.ObjectCannedACLPublicRead), //cho phep public ,xem duoc va download dc
		Bucket:       aws.String(awsBucket),                    //no nam trong bucket gi
		Key:          aws.String(path),                         //voi key gi
		Body:         fileBytes,                                // phan byte cua anh truyen vao body
		CacheControl: aws.String(cacheControl),
		ContentType:  aws.String(mime), //loai anh
	}
	resp, err := svc.PutObject(params) // day putobjecinput len amazone
	if err != nil {
		fmt.Printf("[App.Error] : bad response: %s", err)
		return awsutil.StringValue(resp), err
	}
	//lay link url cua anh tu amazone s3
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(aws_bucket_public_upload_files), //nam o dau
		Key:    aws.String(path),                           //file ten gi
	})
	//lay chuoi url ra
	urlString, err := req.Presign(time.Duration(timeExpries) * time.Minute)

	if err != nil {
		fmt.Printf("[App.Error] : can't get url from amazone s3: %s", err)
		return "", err
	}
	//do url co cac paramater tu amazone dinh kem nen doan code sau se loai bo cac parameter do

	return urlString, nil
}

func uploadFileByEmailAmazon(buf []byte, mime string, file_name string, aws_path string, awsBucket string) (string, error) {

	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_token)
	_, err := creds.Get()
	if err != nil {
		fmt.Printf("[App.Error] : bad credentials: %s", err)
		return "", err
	}
	cfg := aws.NewConfig().WithRegion(aws_region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)
	//tao filebytes tu gia buf truyen vao
	fileBytes := bytes.NewReader(buf)
	//dung dan luu tru anh trong amazon s3
	path := "/" + aws_path + "/" + file_name
	cacheControl := "public, max-age=" + maxAGE
	params := &s3.PutObjectInput{
		ACL:          aws.String(s3.ObjectCannedACLPublicRead), //cho phep public ,xem duoc va download dc
		Bucket:       aws.String(awsBucket),                    //no nam trong bucket gi
		Key:          aws.String(path),                         //voi key gi
		Body:         fileBytes,                                // phan byte cua anh truyen vao body
		CacheControl: aws.String(cacheControl),
		ContentType:  aws.String(mime), //loai anh
	}
	resp, err := svc.PutObject(params) // day putobjecinput len amazone
	if err != nil {
		fmt.Printf("[App.Error] : bad response: %s", err)
		return awsutil.StringValue(resp), err
	}
	//lay link url cua anh tu amazone s3
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(aws_bucket_public_upload_files), //nam o dau
		Key:    aws.String(path),                           //file ten gi
	})
	//lay chuoi url ra
	url_string, err := req.Presign(300 * time.Second)

	if err != nil {
		fmt.Printf("[App.Error] : can't get url from amazone s3: %s", err)
		return "", err
	}
	//do url co cac paramater tu amazone dinh kem nen doan code sau se loai bo cac parameter do
	url, err := url.Parse(url_string)
	if err != nil {
		fmt.Printf("[App.Error] : can't parse string to url: %s", err)
		return "", err
	}
	new_url := url.Host + url.Path
	return new_url, nil
}
