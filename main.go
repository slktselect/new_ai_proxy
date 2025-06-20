package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

var (
	socks5Proxy   = os.Getenv("SOCKS5_PROXY")    // 代理地址，例如 127.0.0.1:1080
	targetBaseURL = os.Getenv("TARGET_BASE_URL") // 目标服务基础地址，例如 https://api.openai.com
)

func main() {
	if socks5Proxy == "" || targetBaseURL == "" {
		log.Fatal("请设置 SOCKS5_PROXY 和 TARGET_BASE_URL 环境变量")
	}

	// 创建 SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", socks5Proxy, nil, proxy.Direct)
	if err != nil {
		log.Fatalf("创建 SOCKS5 dialer 失败: %v", err)
	}
	// 包装 DialContext，忽略 context，直接调用 dialer.Dial
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}
	// 自定义 Transport，所有请求走 SOCKS5
	httpTransport := &http.Transport{
		DialContext:         dialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{Transport: httpTransport}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 拼接目标URL
		path := r.URL.Path
		if strings.HasSuffix(targetBaseURL, "/") {
			path = strings.TrimPrefix(path, "/")
		}
		targetURL := targetBaseURL + path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		log.Printf("[代理请求] %s %s", r.Method, targetURL)

		// 复制请求体
		req, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, "请求构造失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 复制头部，删除 Host
		req.Header = r.Header.Clone()
		req.Header.Del("Host")

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "代理请求失败: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 拷贝响应头
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// 直接流式传输响应体，支持 SSE 流
		io.Copy(w, resp.Body)
	})

	log.Println("代理服务启动在 :11434")
	log.Fatal(http.ListenAndServe(":11434", nil))
}
