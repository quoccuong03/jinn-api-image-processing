package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"encoding/json"

	"bytes"
	"mime/multipart"

)

func TestIndex(t *testing.T) {
	ts := testServer(indexController)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(body), "imaginary") == false {
		t.Fatalf("Invalid body response: %s", body)
	}
}

func TestCrop(t *testing.T) {
	ts := testServer(controller(Crop))
	buf := readFile("imaginary.jpg")
	url := ts.URL
	defer ts.Close()
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, buf)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 550, 740)
	if err != nil {
		t.Error(err)
	}


}

func TestResize(t *testing.T) {
	ts := testServer(controller(Resize))
	buf := readFile("imaginary.jpg")
	url := ts.URL + "?width=300&height=300"
	defer ts.Close()
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, buf)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 300)
	if err != nil {
		t.Error(err)
	}


}

func TestEnlarge(t *testing.T) {
	ts := testServer(controller(Enlarge))
	buf := readFile("imaginary.jpg")
	url := ts.URL + "?width=300&height=400"
	defer ts.Close()
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, buf)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 400)
	if err != nil {
		t.Error(err)
	}

}
func TestProfile(t *testing.T) {
	ts := testServer(controller(Profile))
	buf := readFile("imaginary.jpg")
	url := ts.URL + "?width=300&height=400"
	defer ts.Close()
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, buf)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJob19xdW9jX2N1b25nIiwiZ3JvdXBfaWRzIjpbXSwic2NvcGVzIjpbIldFQiJdLCJ1c2VySWQiOjMyLCJncm91cElkIjpudWxsLCJwcm9maWxlIjp7InVzZXJBdmF0YXIiOnsidXJsIjoiaHR0cHM6Ly9pbWFnZXMuamlubi52bi9hdmF0YXIzMiIsIndpZHRoIjoxMzUsImhlaWdodCI6OTB9LCJmaXJzdE5hbWUiOiJxdW9jIGN1b25nIiwibGFzdE5hbWUiOiJobyJ9LCJyb2xlSWQiOjEsImlzQXBwcm92ZWQiOjAsImVtYWlsIjoiY3VvbmdfdmlwNjVAeWFob28uY29tIiwibGFzdExvZ2luIjp7ImRhdGUiOiIyMDE3LTEwLTA5IDA4OjUxOjEwLjA2ODE5NSIsInRpbWV6b25lX3R5cGUiOjMsInRpbWV6b25lIjoiQXNpYS9Ib19DaGlfTWluaCJ9LCJleHAiOiIxNTA3NTE0MTcwIiwiaWF0IjoxNTA3NTEzODcwfQ.SEIF06biVK7WyrCv3q_JtWF7uJURO19L7Vn2TMhDZqEU-hAyK7Y5O818_suu6IHbJ81mngPmCQxh08VEUAKvnoa8HSgvpiGCcQ73wcnvIo80zewLiwdmpecsrSWO4r6X3y-_CDm1Vatv7E8LCI8P3VlGGQackN2Zxbo9j6ky9Oy8lEkzoDBztvC1y32G2LAiYUEseJSPkaOhS1Tkf44_Yi4qQQiBM0ywi_cUdyYLfGgLvyP6nw05htZWbGoOux_RDm9xUjfTiGiOYDlDqq8KNKZTaVYQPT61Kp7AxP3M5WkCv-umRVIv_sE3WwXhF2-RqcQlQWYpE47iMks-WhWJSHqHq-dM_tekrMpmIRmx_oyo1FWCw0bYOO5WmiEaZtvlexFIhCDYWKgkxaYBGm4B_iNyUS4Y5vATe5_Qd12J350vv_9GJmO1Cq3sGEpU-zvzdQiLIzVkdvSjcZ_ryOPHGO3_Lxc9h0cIvhdIoihOxvJttUcnAUYxvYtae_hq2mGyTPrVhbOs2yKhoLjwAtocUR9SLqX90nulqoc9phpgtUKtIsEuVnVWy5pzeMaVN_OfrvO4zPRujqtwFXcN-WpUcJ3917XoNwRO1zNQyaWSGyN9ZydRKK57I4Hd9ya-6-b4LW39qsHOU0UVs_hxrQKkII55zk30__maqRYbLZfTLjg")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 400)
	if err != nil {
		t.Error(err)
	}

}

func TestUpLoadFile(t *testing.T) {
	ts := testServer(uploadFileController)
	defer ts.Close()
	data,contentType:=makeFileUpload()
	client := &http.Client{}

	req, err := http.NewRequest("POST", ts.URL, data)

	req.Header.Add("Content-Type",contentType)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}




}
func makeFileUpload()(*bytes.Buffer,string){
	file, err := os.Open("hinh_PDF.pdf")
	if err != nil {
		fmt.Println("error opening file")

	}
	defer file.Close()
	data := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(data)
	fileWriter, err := bodyWriter.CreateFormFile("file", "testPDF")
	if err != nil {
		fmt.Println("error writing to buffer")

	}
	_, err = io.Copy(fileWriter, file)

	if err != nil {
		fmt.Println("copying fileWriter %v", err)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	return data,contentType
}
func TestAttachments(t *testing.T) {
	ts := testServer(attachmentsController)
	defer ts.Close()

	client := &http.Client{}
	data,contentType:=makeFileUpload()
	req, err := http.NewRequest("POST", ts.URL+"?groupId=20", data)

	req.Header.Add("Content-Type",contentType)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FHRU5DWV9SRVBSRVNFTlRBVElWRSIsIlJPTEVfVVNFUiJdLCJ1c2VybmFtZSI6ImRhbmhvYW5nQGppbm4udm4iLCJncm91cF9pZHMiOltdLCJzY29wZXMiOlsiV0VCIl0sInVzZXJJZCI6MzI4LCJncm91cElkIjoyMCwicHJvZmlsZSI6eyJ1c2VyQXZhdGFyIjp7InVybCI6bnVsbCwid2lkdGgiOm51bGwsImhlaWdodCI6bnVsbH0sImZpcnN0TmFtZSI6ImRhbiIsImxhc3ROYW1lIjoibmd1eWVuIn0sInJvbGVJZCI6NywiaXNBcHByb3ZlZCI6MCwiZW1haWwiOiJkYW5ob2FuZ0BqaW5uLnZuIiwibGFzdExvZ2luIjp7ImRhdGUiOiIyMDE3LTEwLTExIDEyOjQ5OjI3LjQzODkzMCIsInRpbWV6b25lX3R5cGUiOjMsInRpbWV6b25lIjoiQXNpYS9Ib19DaGlfTWluaCJ9LCJleHAiOiIxNTA3NzAxMjY3IiwiaWF0IjoxNTA3NzAwOTY3fQ.CwSZtX3xOp_StrHMrM35zwtc6360rreWW66yMt9VxsiTxCtDTkAoT9vYBwA0epTHXM6lg3j2FeihubzyWOFMr9A7KCAswwQuvO6G30qi454-E4RTnVQVjsatWziXohERbFqoB9yUDcWMsMOt9qEpPPh1j6Xeck4PuHPzT_lkMZ2Qwhthd1ZQecb81hve9H46e6kYK8PwKkkplS1VObNBbjUpBXUvn4ueCK-fxnpqM3NtwABm7W6QrlXC93nwkt01bnwuXku8LJXvm3GkbG_Npq17C1yprJH8hamMekOQoHkV1sYISiZ_Qt1spkQVofv2jMFqeCWSPO_4Zf0fj8vDKSOIntA7EJ6cVMwF_SPax13PbxEx2GATao_ECfm57tXB7x37-Gd3dp-qk4jRffM2xEVK8bW88FQ7k6d3z3aHiYTW1g2YHye2iMdGQnKoxvL4ehzjLFGDgs0xOsotzbPONSLD8hWspkvb-CHffmkt0e_GgvyLd5LW-jCUlLptLv0XJ_4Ulxdk9cdP7KXNeIyKQSP6-Ql3Yx5psMTMcg_pm4WezBkhAWJ5zb2RXC7k2MUy1s0U-YyzU_NVv2dP95H8FZwt1PMDX7uXSQ8_cpOVFrRvYReNh2_oOtY1hAv8TxEWZO5ybM2Ujw5SQkm8cEU4R63hG5Fyo-1zk3GprTIpnNw")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

}

func TestAttachmentsWithRoleAdmin(t *testing.T) {
	ts := testServer(attachmentsController)
	defer ts.Close()

	client := &http.Client{}
	data,contentType:=makeFileUpload()
	req, err := http.NewRequest("POST", ts.URL, data)

	req.Header.Add("Content-Type",contentType)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX1NVUEVSX0FETUlOIiwiUk9MRV9VU0VSIl0sInVzZXJuYW1lIjoiYWRtaW4iLCJncm91cF9pZHMiOltdLCJzY29wZXMiOlsiV0VCIl0sInVzZXJJZCI6MiwiZ3JvdXBJZCI6bnVsbCwicHJvZmlsZSI6eyJ1c2VyQXZhdGFyIjp7InVybCI6Imh0dHBzOi8vaW1hZ2VzLmppbm4udm4vYmxvYmI2ajhuN2I5N2VnZzAwYzQ2azAwLnBuZyIsIndpZHRoIjoyNTAsImhlaWdodCI6MjUwfSwiZmlyc3ROYW1lIjoiZGFiIiwibGFzdE5hbWUiOiJhc2VhZWUifSwicm9sZUlkIjo5LCJpc0FwcHJvdmVkIjpudWxsLCJlbWFpbCI6ImFkbWluQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMTEgMTM6MTI6MzcuMTUzNzUzIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDc3MDI2NTciLCJpYXQiOjE1MDc3MDIzNTd9.kGWLUnXHCN9P-_boFP340c0lV6fJDN08BucKJD7uS8sXfFU9FioBEi0Bz3Rub9qYHOfdt9x7tQmx6w9ejesraoOZFInk2qe7tNVeNU2TciE9ESh-I3Bbs2KFC73aSrtSuOhbXRmaqDDPmZeRBRl2JPsyy3yvE3dLjJ7ubpM__RUV3-IjI9XP8RBxgx46k_kZR_ahq_KJjpZiO0kxnSZ8yGnuBPagaXLN6BYbIg6z_CT_-h_3CmikFp61xGeGtx37qqFVXLCRYC-oNdMYpp9vN4lvRUHLLQS0U1HujhGiG7NHRl4L5g6LN_S3iyKivPWBaayrsH6788XA3QdxDJBzU8G7jx6hr3bsXSZPfl0slWMdCBSzlN88ZaFzuyLtY-R1S-lwzZd8R4vW1kucFz85humKHWLy9aoT-fwl8xhel-KEe9H0bpM_Xx3eETCp448aio301uVtP_ZWcrK1Nfb-Vx01Uq92k7RzIGyZKw4_vz8OPCmnwzeHpuJ6zk5dmCQ-uXaI_4etdRXvRVNXAraX6jiuJV8dVx1X8zpn4A4NzdSEO-Rnp_RDH3xtWI5rb2Ub1i72OLh7G3NQyWz5Z8DXr_na2Ze6vum__Y-zCCa0RWqfCfKlZoi8wqtEv5Rc0E7tSiJvKBvrwZ7xFnpow2QjNOXVTNMJnZRh5NajKX9F2IE")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

}
func TestAttachmentsInvalidFilter(t *testing.T) {
	ts := testServer(attachmentsController)
	defer ts.Close()

	client := &http.Client{}
	data,contentType:=makeFileUpload()
	req, err := http.NewRequest("POST", ts.URL+"?groupId=2", data)

	req.Header.Add("Content-Type",contentType)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FHRU5DWV9SRVBSRVNFTlRBVElWRSIsIlJPTEVfVVNFUiJdLCJ1c2VybmFtZSI6ImRhbmhvYW5nQGppbm4udm4iLCJncm91cF9pZHMiOltdLCJzY29wZXMiOlsiV0VCIl0sInVzZXJJZCI6MzI4LCJncm91cElkIjoyMCwicHJvZmlsZSI6eyJ1c2VyQXZhdGFyIjp7InVybCI6bnVsbCwid2lkdGgiOm51bGwsImhlaWdodCI6bnVsbH0sImZpcnN0TmFtZSI6ImRhbiIsImxhc3ROYW1lIjoibmd1eWVuIn0sInJvbGVJZCI6NywiaXNBcHByb3ZlZCI6MCwiZW1haWwiOiJkYW5ob2FuZ0BqaW5uLnZuIiwibGFzdExvZ2luIjp7ImRhdGUiOiIyMDE3LTEwLTExIDEyOjQ5OjI3LjQzODkzMCIsInRpbWV6b25lX3R5cGUiOjMsInRpbWV6b25lIjoiQXNpYS9Ib19DaGlfTWluaCJ9LCJleHAiOiIxNTA3NzAxMjY3IiwiaWF0IjoxNTA3NzAwOTY3fQ.CwSZtX3xOp_StrHMrM35zwtc6360rreWW66yMt9VxsiTxCtDTkAoT9vYBwA0epTHXM6lg3j2FeihubzyWOFMr9A7KCAswwQuvO6G30qi454-E4RTnVQVjsatWziXohERbFqoB9yUDcWMsMOt9qEpPPh1j6Xeck4PuHPzT_lkMZ2Qwhthd1ZQecb81hve9H46e6kYK8PwKkkplS1VObNBbjUpBXUvn4ueCK-fxnpqM3NtwABm7W6QrlXC93nwkt01bnwuXku8LJXvm3GkbG_Npq17C1yprJH8hamMekOQoHkV1sYISiZ_Qt1spkQVofv2jMFqeCWSPO_4Zf0fj8vDKSOIntA7EJ6cVMwF_SPax13PbxEx2GATao_ECfm57tXB7x37-Gd3dp-qk4jRffM2xEVK8bW88FQ7k6d3z3aHiYTW1g2YHye2iMdGQnKoxvL4ehzjLFGDgs0xOsotzbPONSLD8hWspkvb-CHffmkt0e_GgvyLd5LW-jCUlLptLv0XJ_4Ulxdk9cdP7KXNeIyKQSP6-Ql3Yx5psMTMcg_pm4WezBkhAWJ5zb2RXC7k2MUy1s0U-YyzU_NVv2dP95H8FZwt1PMDX7uXSQ8_cpOVFrRvYReNh2_oOtY1hAv8TxEWZO5ybM2Ujw5SQkm8cEU4R63hG5Fyo-1zk3GprTIpnNw")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 400 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

}

//func TestExtract(t *testing.T) {
//	ts := testServer(controller(Extract))
//	buf, _ := ioutil.ReadAll(readFile("imaginary.jpg"))
//	urls := ts.URL
//	defer ts.Close()
//	client := &http.Client{}
//	param := url.Values{}
//
//	dulieu:=`{"areawidth":`+`200`+`,`+`"areaheight":`+`200}`
//	param.Add("file",string(buf))
//	param.Add("data",dulieu)
//	req, err := http.NewRequest("POST", urls, bytes.NewBufferString(param.Encode()))
//	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
//	res, err := client.Do(req)
//	defer res.Body.Close()
//	if err != nil {
//		t.Fatal("Cannot perform the request")
//	}
//
//	if res.StatusCode != 200 {
//		t.Fatalf("Invalid response status: %s", res.Status)
//	}
//
//	image, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if len(image) == 0 {
//		t.Fatalf("Empty response body")
//	}
//
//	err = assertSize(image, 200, 120)
//	if err != nil {
//		t.Error(err)
//	}
//
//
//}

func TestRemoteHTTPSource(t *testing.T) {
	opts := ServerOptions{EnableURLSource: true}
	fn := ImageMiddleware(opts)(Crop)
	LoadSources(opts)

	tsImage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadFile("fixtures/large.jpg")
		w.Write(buf)
	}))
	defer tsImage.Close()

	ts := httptest.NewServer(fn)
	url := ts.URL + "?width=200&height=200&url=" + tsImage.URL
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url,nil)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %d", res.StatusCode)
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 200, 200)
	if err != nil {
		t.Error(err)
	}

}

func TestInvalidRemoteHTTPSource(t *testing.T) {
	opts := ServerOptions{EnableURLSource: true}
	fn := ImageMiddleware(opts)(Crop)
	LoadSources(opts)

	tsImage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(400)
	}))
	defer tsImage.Close()

	ts := httptest.NewServer(fn)
	url := ts.URL + "?width=200&height=200&url=" + tsImage.URL
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url,nil)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Request failed")
	}
	if res.StatusCode != 400 {
		t.Fatalf("Invalid response status: %d", res.StatusCode)
	}
}

func TestMountDirectory(t *testing.T) {
	opts := ServerOptions{Mount: "fixtures"}
	fn := ImageMiddleware(opts)(Crop)
	LoadSources(opts)

	ts := httptest.NewServer(fn)
	url := ts.URL + "?width=200&height=200&file=large.jpg"
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url,nil)
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJyb2xlcyI6WyJST0xFX0FDQ09VTlRBTlQiLCJST0xFX1VTRVIiXSwidXNlcm5hbWUiOiJrZXR0b2FuZCIsImdyb3VwX2lkcyI6W10sInNjb3BlcyI6WyJBRE1JTiJdLCJ1c2VySWQiOjg2OSwiZ3JvdXBJZCI6NjAsInByb2ZpbGUiOnsidXNlckF2YXRhciI6eyJ1cmwiOm51bGwsIndpZHRoIjowLCJoZWlnaHQiOjB9LCJmaXJzdE5hbWUiOiJkIiwibGFzdE5hbWUiOiJrZXRvYW4ifSwicm9sZUlkIjozLCJpc0FwcHJvdmVkIjowLCJlbWFpbCI6ImtldHRvYW5kQGppbm4udm4iLCJsYXN0TG9naW4iOnsiZGF0ZSI6IjIwMTctMTAtMDcgMDg6NTg6NTcuODI2NzIwIiwidGltZXpvbmVfdHlwZSI6MywidGltZXpvbmUiOiJBc2lhL0hvX0NoaV9NaW5oIn0sImV4cCI6IjE1MDczNDE4MzciLCJpYXQiOjE1MDczNDE1Mzd9.oKZgGyXreP9WvGIBKqQsbjAr1v_4JOLfq9pbonUsUzAt4Ik1QUFJb-p_O5A16w5k5s2puP_ja8YKn31y55cJ3-YupBnlH2GD82xrepIkiN570Z2TQMh_bwM3Rp0dAJKnBz62BbkQBktHEatNunrYyPg3XHRAAhses-xmVKur_zAnB7RozonbzANgeHl6DfXT352ECPjhD5PCkUN5JUQnSZSF_OUfPnvsx2xDkOQZyNO7Wir7g8M-g_Y0EOSYav8LYqmt__CU5992qjT-371FXMezm8ax9Fodx3iAtnIISlFpUbq_zaV6QeNIOOJzuJEOb9p2Ahf5rYMl272dp5-SJgDedXyTXi6eGAjojtyevFnEkJa42llrB8zsoBWoCZAF6OJ-ZUhYykxups_TATDC5QHmJyk2xTbRGENBCG1venlB_CtRJ3dAA7Ckg8hegpqSYXfiB-L7ufzdiwLrq5uwbTD1QanFI8-9DwNVLYsOvbALW5e3QYOSl4VH9ldA62gMAOFR9i7b7fe0990elLAIgxVcCEh8vsv62jpVAxWafd80ZXbF5X9TqSJNC6jNwAMI7AlA1hDAOYcGV3tWHjU_QfcXamM3yhIFfXl2rPGCuc86T12ORHpHTsI8xE-j7iwnmJSdav_HVUtFN_IVlqQ3lwxTZb9kIUo2687lQOxUPIY")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %d", res.StatusCode)
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 200, 200)
	if err != nil {
		t.Error(err)
	}


}


func TestMountInvalidPath(t *testing.T) {
	fn := ImageMiddleware(ServerOptions{Mount: "_invalid_"})(Crop)
	ts := httptest.NewServer(fn)
	url := ts.URL + "?top=100&left=100&areawidth=200&areaheight=120&file=../../large.jpg"
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}

	if res.StatusCode != 400 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}
}

func controller(op Operation) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		buf, _ := ioutil.ReadAll(r.Body)
		imageHandler(w, r, buf, op, ServerOptions{})
	}
}

func testServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func readFile(file string) io.Reader {
	buf, _ := os.Open(path.Join("fixtures", file))
	return buf
}

func assertSize(buf []byte, width, height int) error {
	var info ImageImfomation
	err:=json.Unmarshal(buf, &info)
	if err != nil {
		return err
	}
	if info.OriginalWidth != width || info.Originalheight != height {
		return fmt.Errorf("Invalid image size: %dx%d", info.OriginalWidth, info.Originalheight)
	}
	return nil
}

