package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/daaku/go.httpgzip"
	"github.com/rs/cors"
	"gopkg.in/h2non/bimg.v1"
	"gopkg.in/throttled/throttled.v2"
	"gopkg.in/throttled/throttled.v2/store/memstore"

	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/dgrijalva/jwt-go"

	"strconv"
)

type checkRoles struct {
	Roles   []string `json:"roles"`
	UserId  int      `json:"userId"`
	GroupId int      `json:"groupId"`
}

func Middleware(fn func(http.ResponseWriter, *http.Request), o ServerOptions) http.Handler {
	next := http.Handler(http.HandlerFunc(fn))

	if o.Concurrency > 0 {
		next = throttle(next, o)
	}
	if o.Gzip {
		next = httpgzip.NewHandler(next)
	}
	if o.CORS {

		next = cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler(next)
	}

	if o.ApiKey != "" {
		next = authorizeClient(next, o)
	}
	if o.HttpCacheTtl >= 0 {
		next = setCacheHeaders(next, o.HttpCacheTtl)
	}

	return validate(defaultHeaders(next), o)
}

func UnAuthMiddleware(fn func(http.ResponseWriter, *http.Request), o ServerOptions) http.Handler {
	next := http.Handler(http.HandlerFunc(fn))

	if o.Concurrency > 0 {
		next = throttle(next, o)
	}
	if o.Gzip {
		next = httpgzip.NewHandler(next)
	}
	if o.CORS {

		next = cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler(next)
	}

	if o.HttpCacheTtl >= 0 {
		next = setCacheHeaders(next, o.HttpCacheTtl)
	}

	return validate(defaultHeaders(next), o)
}

func ImageMiddleware(o ServerOptions) func(Operation) http.Handler {
	return func(fn Operation) http.Handler {
		return validateImage(Middleware(imageController(o, Operation(fn)), o), o)
	}
}

func throttleError(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "throttle error: "+err.Error(), http.StatusInternalServerError)
	})
}

func throttle(next http.Handler, o ServerOptions) http.Handler {
	store, err := memstore.New(65536)
	if err != nil {
		return throttleError(err)
	}

	quota := throttled.RateQuota{throttled.PerSec(o.Concurrency), o.Burst}
	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		return throttleError(err)
	}

	httpRateLimiter := throttled.HTTPRateLimiter{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Method: true},
	}

	return httpRateLimiter.RateLimit(next)
}

func validate(next http.Handler, o ServerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "POST" && r.Method != "OPTIONS" {
			ErrorReply(r, w, ErrMethodNotAllowed, o)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateImage(next http.Handler, o ServerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _, _ := r.FormFile("file")
		if file != nil {
			if file.(Sizer).Size() > (10 * MB_FILE) {
				ErrorReply(r, w, ErrEntityTooLarge, o)
				return
			}
		}

		path := r.URL.Path
		if r.Method == "GET" && isPublicPath(path) {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method == "GET" && o.Mount == "" && o.EnableURLSource == false {
			ErrorReply(r, w, ErrMethodNotAllowed, o)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authorizeClient(next http.Handler, o ServerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//neu mehthod la options thi bo qua kiem tra token

		if r.Method != "OPTIONS" {

			jwt_token, err := newToken(w, r)

			if err == nil && jwt_token.Valid {
				// convert type claims thanh mot mang chua json cua token
				data := jwt_token.Claims.(jwt.MapClaims)

				if data["roles"] == nil {
					ErrorReply(r, w, ErrForbidden, o)
					return
				} else {
					//vi data la kieu interface {} khong sua dung duoc voi for nen ta se encode
					roles, err := json.Marshal(data)

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
					}
					//tao bien count la 0 neu count 1 la thoa dk
					count := 0
					for _, r := range checkroles.Roles {
						for _, arr := range array_roles {
							if r == arr {
								count = 1
								break
							}
						}
						if count == 1 {
							break
						}
					}

					if count == 0 {
						ErrorReply(r, w, ErrForbidden, o)
						return
					}

				}

			} else {
				ErrorReply(r, w, ErrInvalidApiKey, o)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func getGroupId(w http.ResponseWriter, r *http.Request) (string, error) {
	jwt_token, err := newToken(w, r)
	if err == nil && jwt_token.Valid {
		// convert type claims thanh mot mang chua json cua token
		data := jwt_token.Claims.(jwt.MapClaims)
		roles, err := json.Marshal(data)
		if err != nil {
			log.Printf("[App.Error: can't marshal data in getGroupId %s ", err)
			return "", err
		}

		//tiep theo tao mot interface roles moi va decoder du lieu vao
		var checkroles checkRoles
		err = json.Unmarshal(roles, &checkroles)

		if err != nil {
			log.Printf("[App.Error: can't marshal data in getGroupId %s ", err)
			return "", err
		}
		if checkAdminUser(checkroles) == true {
			return "admin", nil
		} else if strconv.Itoa(checkroles.GroupId) == "" {
			return "", ErrRoleUserNotAllowed
		}
		return strconv.Itoa(checkroles.GroupId), err

	} else {
		log.Printf("[AppError] can't creat Token in getGroupId")
		return "", err
	}

}
func checkAdminUser(checkroles checkRoles) bool {
	count := 0
	for _, r := range checkroles.Roles {
		for _, arr := range array_admin {
			if r == arr {
				count = 1
				break
			}
		}
		if count == 1 {
			break
		}
	}

	if count == 0 {
		return false
	} else {
		return true
	}
}

func defaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", fmt.Sprintf("imaginary %s (bimg %s)", Version, bimg.Version))

		next.ServeHTTP(w, r)
	})
}

func setCacheHeaders(next http.Handler, ttl int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer next.ServeHTTP(w, r)

		if r.Method != "GET" || isPublicPath(r.URL.Path) {
			return
		}

		ttlDiff := time.Duration(ttl) * time.Second
		expires := time.Now().Add(ttlDiff)

		w.Header().Add("Expires", strings.Replace(expires.Format(time.RFC1123), "UTC", "GMT", -1))
		w.Header().Add("Cache-Control", getCacheControl(ttl))
	})
}

func getCacheControl(ttl int) string {
	if ttl == 0 {
		return "private, no-cache, no-store, must-revalidate"
	}
	return fmt.Sprintf("public, s-maxage=%d, max-age=%d, no-transform", ttl, ttl)
}

func isPublicPath(path string) bool {
	return path == "/" || path == "/health" || path == "/form"
}

func newToken(w http.ResponseWriter, r *http.Request) (*jwt.Token, error) {
	//lay string token tu authorization
	key := r.Header.Get("Authorization")
	//neu chieu dai cua chuoi author be hon 8 ky tu thi bao error
	if len(key) < 8 {

		return nil, ErrInvalidApiKey
	}
	//cat bo Bearer
	key = key[7:len(key)]

	//lay token ra
	jwt_token, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {
		//chuyen serect key ve []byte
		publicKey, err := ioutil.ReadFile(public_key)
		if err != nil {
			return nil, fmt.Errorf("Error reading public key")
		}
		//ma hoa serect key
		new_publickey, error := jwt.ParseRSAPublicKeyFromPEM(publicKey)
		if error != nil {
			return nil, fmt.Errorf("Error ParseRSA publickey ")
		}

		return new_publickey, nil
	})
	return jwt_token, err

}
