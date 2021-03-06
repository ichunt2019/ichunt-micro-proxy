package ichunt_micro_proxy

import (
	"context"
	"fmt"
	"github.com/ichunt2019/ichunt-micro-proxy/proxy/load_balance"
	"github.com/ichunt2019/ichunt-micro-registry/registry"
	_ "github.com/ichunt2019/ichunt-micro-registry/registry/etcd"
	"log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

type RealServer struct {
	Addr string
}


func TestRegister(t *testing.T) {
	fmt.Println("66666")
	rs2 := &RealServer{Addr: "192.168.2.246:2004"}
	rs2.Run()

	//服务注册
	register()


	//监听关闭信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}


func register(){
	registryInst, err := registry.InitRegistry(context.TODO(), "etcd",
		registry.WithAddrs([]string{"192.168.2.232:2379"}),
		registry.WithTimeout(time.Second),
		registry.WithPasswrod(""),
		registry.WithRegistryPath("/ichuntMicroService/"),
		registry.WithHeartBeat(5),
	)



	if err != nil {
		fmt.Printf("init registry failed, err:%v", err)
		return
	}

	load_balance.Init(registryInst)

	service := &registry.Service{
		Name: "comment_service",
	}

	service.Nodes = append(service.Nodes,
		//&registry.Node{
		//	IP:   "192.168.2.232",
		//	Port: 2003,
		//	Weight:1,
		//},
		&registry.Node{
			IP:   "192.168.2.246",
			Port: 2004,
			Weight:2,
		},
	)
	registryInst.Register(context.TODO(), service)
}

func (r *RealServer) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", r.HelloHandler)
	mux.HandleFunc("/base/error", r.ErrorHandler)
	server := &http.Server{
		Addr:         r.Addr,
		WriteTimeout: time.Second * 3,
		Handler:      mux,
	}
	go func() {
		log.Println(server.ListenAndServe())
	}()
}

func (r *RealServer) HelloHandler(w http.ResponseWriter, req *http.Request) {
	//127.0.0.1:8008/abc?sdsdsa=11
	//r.Addr=127.0.0.1:8008
	//req.URL.Path=/abc
	//time.Sleep(time.Second)
	fmt.Println("host:",req.Host)
	fmt.Println("header:",req.Header)
	fmt.Println("cookie:",req.Cookies())
	fmt.Println(req.ParseForm())
	fmt.Println("post params: ",req.PostForm)
	fmt.Println("url :",req.URL)
	fmt.Println("url rawpath :",req.URL.RawPath)
	fmt.Println("query :",req.URL.Query())

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("read body err, %v\n", err)
		return
	}
	println("json:", string(body))

	upath := fmt.Sprintf("http://%s%s\n", r.Addr, req.URL.Path)
	realIP := fmt.Sprintf("RemoteAddr=%s,X-Forwarded-For=%v,X-Real-Ip=%v\n", req.RemoteAddr, req.Header.Get("X-Forwarded-For"),
		req.Header.Get("X-Real-Ip"))
	io.WriteString(w, upath)
	io.WriteString(w, realIP)
}

func (r *RealServer) ErrorHandler(w http.ResponseWriter, req *http.Request) {
	upath := "error handler"
	w.WriteHeader(500)
	io.WriteString(w, upath)
}