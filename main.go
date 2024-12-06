package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type Respond struct {
	Success bool    `json:"success"`
	Status  string  `json:"status"`
	Data    []Proxy `json:"data"`
}

type Proxy struct {
	Remark        string `json:"remark"`        // 描述
	Prefix        string `json:"prefix"`        // 转发的前缀判断
	Upstream      string `json:"upstream"`      // 后端nginx或ip地址
	RewritePrefix string `json:"rewritePrefix"` // 重写
}

var (
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	proxyMap = make(map[string]Proxy)
)

var adminUrl = flag.String("adminUrl", "", "admin 的地址")
var profile = flag.String("profile", "", "环境")
var proxyFile = flag.String("proxyFile", "", "测试环境的数据")

// 初始化日志
func initLog() {
	fmt.Println("加载配置")
	errFile, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	infoFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	InfoLog = log.New(io.MultiWriter(os.Stderr, infoFile), "Info:", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	ErrorLog = log.New(io.MultiWriter(os.Stderr, errFile), "Error:", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
}

func initProxyList() {
	resp, _ := http.Get(*adminUrl)
	if resp != nil && resp.StatusCode == 200 {
		bytes, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			fmt.Println("ioutil.ReadAll err=", err)
			return
		}
		var respond Respond
		err = json.Unmarshal(bytes, &respond)
		if err != nil {
			fmt.Println("json.Unmarshal err=", err)
			return
		}
		proxyList := respond.Data
		for _, proxy := range proxyList {
			proxyMap[proxy.Prefix+"/"] = proxy
		}
	}
}

func main() {
	router := gin.Default() //创建一个router
	flag.Parse()
	initLog()
	if *profile != "" {
		InfoLog.Printf("加载远端数据: %s ", *adminUrl)
		initProxyList()
	} else {
		InfoLog.Printf("加载本地配置数据: %s", *proxyFile)
		loadProxyListFromFile()
	}
	router.Any("/*action", Forward) //所有请求都会经过Forward函数转发

	router.Run(":8000")
}

func Forward(c *gin.Context) {
	HostReverseProxy(c.Writer, c.Request)
}

func HostReverseProxy(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		io.WriteString(w, "Request path Error")
		return
	}

	// 从内存里面获取转发的url
	var upstream = ""
	for prefix, proxy := range proxyMap {
		if strings.HasPrefix(r.URL.Path, prefix) {
			upstream = proxy.Upstream
			rewritePrefix := proxy.RewritePrefix

			// 如果转发的地址是 / 结尾，需要过滤掉
			if strings.HasSuffix(upstream, "/") {
				upstream = strings.TrimRight(upstream, "/")
			}

			// 如果 rewritePrefix 不为空，替换原路径的前缀
			if rewritePrefix != "" {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
				if !strings.HasPrefix(rewritePrefix, "/") {
					rewritePrefix = "/" + rewritePrefix
				}
				r.URL.Path = rewritePrefix + r.URL.Path
			} else {
				// 如果 rewritePrefix 为空，则保持原来的路径
				r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			}

			break
		}
	}

	if upstream == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	remote, err := url.Parse(upstream)
	InfoLog.Printf("RequestURI %s upstream %s remote %s", r.RequestURI, upstream, remote.String())
	if err != nil {
		http.Error(w, "Invalid upstream URL", http.StatusInternalServerError)
		return
	}

	r.URL.Host = remote.Host
	r.URL.Scheme = remote.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = remote.Host

	httputil.NewSingleHostReverseProxy(remote).ServeHTTP(w, r)
}

func loadProxyListFromFile() {
	file, err := os.Open(*proxyFile)
	if err != nil {
		ErrorLog.Println("err:", err)
	}
	var respond Respond
	// 创建json解码器
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&respond)
	if err != nil {
		fmt.Println("LoadProxyListFromFile failed", err.Error())
	}
	proxyList := respond.Data
	for _, proxy := range proxyList {
		proxyMap[proxy.Prefix] = proxy
	}
}
