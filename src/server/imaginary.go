package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	d "runtime/debug"
	"strconv"
	"strings"
	"time"

	bimg "gopkg.in/h2non/bimg.v1"
)

//bien moi truong
var (
	//cac bien congfigure khi ket noi voi amazon
	aws_access_key_id              = envString("AWS_ACCESS_KEY_ID", "AKIAJBZFWM3DOCEYEQTQ")
	aws_secret_access_key          = envString("AWS_SECRET_ACCESS_KEY", "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")
	aws_token                      = envString("AWS_TOKEN", "")
	aws_path_img                   = envString("AWS_PATH_IMG", "image")
	aws_path_file                  = envString("AWS_PATH_FILE", "file")
	aws_path_attachments           = envString("AWS_PATH_ATTACHMENTS", "attachments")
	aws_path_public_upload_files   = envString("AWS_PATH_PUBLIC_UPLOAD_FILES", "publicupload")
	aws_path_upload_email          = envString("AWS_PATH_PUBLIC_UPLOAD_FILES", "upload-by-email")
	aws_bucket                     = envString("AWS_BUCKET", "imaginaryabcd")
	aws_bucket_attachments         = envString("AWS_BUCKET_ATTACHMENTS", "attachmentsfiles")
	aws_bucket_public_upload_files = envString("AWS_BUCKET_PUBLIC_UPLOAD_FILES", "imaginaryabcd")
	aws_region                     = envString("AWS_REGION", "eu-west-2")
	cdn_url                        = envString("CDN_URL", "https://images.jinn.vn")
	//day la time lay out
	time_layout = envString("TIME_LAYOUT", "2006-01-02 15:04:05")
	//day la duong dan cua thu muc chua anh watermark
	image_watermark_path = envString("IMAGE_WATERMARK_PATH", "logo.png")
	public_key           = envString("PUBLIC_KEY", "keys/public.pem")
	CF_DistributionID    = envString("CF_DistributionID", "E1FJTY2UBPJJN1")
	//hostname JinnReal for fromexternalUpload
	hostNameJinnReal = envString("http://jinnreal.com", "hostNameJinnReal")
	MB_FILE          = int64(1000000)
	maxAGE           = envString("MAXAGE", "31557600")
)
var (
	aAddr              = flag.String("a", "", "bind address")
	aPort              = flag.Int("p", 80, "port to listen")
	aVers              = flag.Bool("v", false, "Show version")
	aVersl             = flag.Bool("version", false, "Show version")
	aHelp              = flag.Bool("h", false, "Show help")
	aHelpl             = flag.Bool("help", false, "Show help")
	aPathPrefix        = flag.String("path-prefix", "/", "Url path prefix to listen to")
	aCors              = flag.Bool("cors", true, "Enable CORS support")
	aGzip              = flag.Bool("gzip", false, "Enable gzip compression")
	aAuthForwarding    = flag.Bool("enable-auth-forwarding", false, "Forwards X-Forward-Authorization or Authorization header to the image source server. -enable-url-source flag must be defined. Tip: secure your server from public access to prevent attack vectors")
	aEnableURLSource   = flag.Bool("enable-url-source", false, "Enable remote HTTP URL image source processing")
	aEnablePlaceholder = flag.Bool("enable-placeholder", false, "Enable image response placeholder to be used in case of error")
	aAlloweOrigins     = flag.String("allowed-origins", "", "Restrict remote image source processing to certain origins (separated by commas)")
	aMaxAllowedSize    = flag.Int("max-allowed-size", 0, "Restrict maximum size of http image source (in bytes)")
	aKey               = flag.String("key", "jwt", "Define API key for authorization")
	aMount             = flag.String("mount", "", "Mount server local directory")
	aCertFile          = flag.String("certfile", "", "TLS certificate file path")
	aKeyFile           = flag.String("keyfile", "", "TLS private key file path")
	aAuthorization     = flag.String("authorization", "", "Defines a constant Authorization header value passed to all the image source servers. -enable-url-source flag must be defined. This overwrites authorization headers forwarding behavior via X-Forward-Authorization")
	aPlaceholder       = flag.String("placeholder", "", "Image path to image custom placeholder to be used in case of error. Recommended minimum image size is: 1200x1200")
	aHttpCacheTtl      = flag.Int("http-cache-ttl", -1, "The TTL in seconds")
	aReadTimeout       = flag.Int("http-read-timeout", 60, "HTTP read timeout in seconds")
	aWriteTimeout      = flag.Int("http-write-timeout", 60, "HTTP write timeout in seconds")
	aConcurrency       = flag.Int("concurrency", 0, "Throttle concurrency limit per second")
	aBurst             = flag.Int("burst", 100, "Throttle burst max cache size")
	aMRelease          = flag.Int("mrelease", 30, "OS memory release interval in seconds")
	aCpus              = flag.Int("cpus", runtime.GOMAXPROCS(-1), "Number of cpu cores to use")
)

var array_roles = []string{"ROLE_PROJECT_ADMIN", "ROLE_ACCOUNTANT", "ROLE_MARKETING_HEAD", "ROLE_MARKETING_STAFF", "ROLE_ENCODER", "ROLE_SALE_ADMIN", "ROLE_PROJECT_ADMIN", "ROLE_AGENCY_REPRESENTATIVE", "ROLE_ADMIN", "ROLE_SUPER_ADMIN", "ROLE_SALE"}
var array_admin = []string{"ROLE_ADMIN", "ROLE_SUPER_ADMIN"}

const usage = `imaginary %s

Usage:
  imaginary -p 80
  imaginary -cors -gzip
  imaginary -concurrency 10
  imaginary -path-prefix /api/v1
  imaginary -enable-url-source
  imaginary -enable-url-source -allowed-origins http://localhost,http://server.com
  imaginary -enable-url-source -enable-auth-forwarding
  imaginary -enable-url-source -authorization "Basic AwDJdL2DbwrD=="
  imaginary -enable-placeholder
  imaginary -enable-url-source -placeholder ./placeholder.jpg
  imaginary -h | -help
  imaginary -v | -version

Options:
  -a <addr>                 bind address [default: *]
  -p <port>                 bind port [default: 8088]
  -h, -help                 output help
  -v, -version              output version
  -path-prefix <value>      Url path prefix to listen to [default: "/"]
  -cors                     Enable CORS support [default: false]
  -gzip                     Enable gzip compression [default: false]
  -key <key>                Define API key for authorization
  -mount <path>             Mount server local directory
  -http-cache-ttl <num>     The TTL in seconds. Adds caching headers to locally served files.
  -http-read-timeout <num>  HTTP read timeout in seconds [default: 30]
  -http-write-timeout <num> HTTP write timeout in seconds [default: 30]
  -enable-url-source        Restrict remote image source processing to certain origins (separated by commas)
  -enable-placeholder       Enable image response placeholder to be used in case of error [default: false]
  -enable-auth-forwarding   Forwards X-Forward-Authorization or Authorization header to the image source server. -enable-url-source flag must be defined. Tip: secure your server from public access to prevent attack vectors
  -allowed-origins <urls>   Restrict remote image source processing to certain origins (separated by commas)
  -max-allowed-size <bytes> Restrict maximum size of http image source (in bytes)
  -certfile <path>          TLS certificate file path
  -keyfile <path>           TLS private key file path
  -authorization <value>    Defines a constant Authorization header value passed to all the image source servers. -enable-url-source flag must be defined. This overwrites authorization headers forwarding behavior via X-Forward-Authorization
  -placeholder <path>       Image path to image custom placeholder to be used in case of error. Recommended minimum image size is: 1200x1200
  -concurreny <num>         Throttle concurrency limit per second [default: disabled]
  -burst <num>              Throttle burst max cache size [default: 100]
  -mrelease <num>           OS memory release interval in seconds [default: 30]
  -cpus <num>               Number of used cpu cores.
                            (default for current machine is %d cores)
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, Version, runtime.NumCPU()))
	}
	flag.Parse()

	if *aHelp || *aHelpl {
		showUsage()
	}
	if *aVers || *aVersl {
		showVersion()
	}

	// Only required in Go < 1.5
	runtime.GOMAXPROCS(*aCpus)

	port := getPort(*aPort)
	opts := ServerOptions{
		Port:              port,
		Address:           *aAddr,
		Gzip:              *aGzip,
		CORS:              *aCors,
		AuthForwarding:    *aAuthForwarding,
		EnableURLSource:   *aEnableURLSource,
		EnablePlaceholder: *aEnablePlaceholder,
		PathPrefix:        *aPathPrefix,
		ApiKey:            *aKey,
		Concurrency:       *aConcurrency,
		Burst:             *aBurst,
		Mount:             *aMount,
		CertFile:          *aCertFile,
		KeyFile:           *aKeyFile,
		Placeholder:       *aPlaceholder,
		HttpCacheTtl:      *aHttpCacheTtl,
		HttpReadTimeout:   *aReadTimeout,
		HttpWriteTimeout:  *aWriteTimeout,
		Authorization:     *aAuthorization,
		AlloweOrigins:     parseOrigins(*aAlloweOrigins),
		MaxAllowedSize:    *aMaxAllowedSize,
		CFDistributionId:  CF_DistributionID,
	}

	// Create a memory release goroutine
	if *aMRelease > 0 {
		memoryRelease(*aMRelease)
	}

	// Check if the mount directory exists, if present
	if *aMount != "" {
		checkMountDirectory(*aMount)
	}

	// Validate HTTP cache param, if present
	if *aHttpCacheTtl != -1 {
		checkHttpCacheTtl(*aHttpCacheTtl)
	}

	// Read placeholder image, if required
	if *aPlaceholder != "" {
		buf, err := ioutil.ReadFile(*aPlaceholder)
		if err != nil {
			exitWithError("cannot start the server: %s", err)
		}

		imageType := bimg.DetermineImageType(buf)
		if !bimg.IsImageTypeSupportedByVips(imageType).Load {
			exitWithError("Placeholder image type is not supported. Only JPEG, PNG or WEBP are supported")
		}

		opts.PlaceholderImage = buf
	} else if *aEnablePlaceholder {
		// Expose default placeholder
		opts.PlaceholderImage = placeholder
	}

	log.Println("imaginary server listening on port :%d/%s", opts.Port, strings.TrimPrefix(opts.PathPrefix, "/"))

	// Load image source providers
	LoadSources(opts)

	// Start the server
	err := Server(opts)
	if err != nil {
		exitWithError("cannot start the server: %s", err)
	}
}

func getPort(port int) int {
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		newPort, _ := strconv.Atoi(portEnv)
		if newPort > 0 {
			port = newPort
		}
	}
	return port
}

func showUsage() {
	flag.Usage()
	os.Exit(1)
}

func showVersion() {
	fmt.Println(Version)
	os.Exit(1)
}

func checkMountDirectory(path string) {
	src, err := os.Stat(path)
	if err != nil {
		exitWithError("error while mounting directory: %s", err)
	}
	if src.IsDir() == false {
		exitWithError("mount path is not a directory: %s", path)
	}
	if path == "/" {
		exitWithError("cannot mount root directory for security reasons")
	}
}

func checkHttpCacheTtl(ttl int) {
	if ttl < -1 || ttl > 31556926 {
		exitWithError("The -http-cache-ttl flag only accepts a value from 0 to 31556926")
	}

	if ttl == 0 {
		log.Println("Adding HTTP cache control headers set to prevent caching.")
	}
}

func parseOrigins(origins string) []*url.URL {
	urls := []*url.URL{}
	if origins == "" {
		return urls
	}
	for _, origin := range strings.Split(origins, ",") {
		u, err := url.Parse(origin)
		if err != nil {
			continue
		}
		urls = append(urls, u)
	}
	return urls
}

func memoryRelease(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for range ticker.C {
			log.Println("FreeOSMemory()")
			d.FreeOSMemory()
		}
	}()
}

func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args)
}

/**
* Load environment
 */
func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
