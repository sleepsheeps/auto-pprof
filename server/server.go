package server

import (
	"auto-pprof/config"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/google/pprof/bridge"
)

// 根据使用者的请求自动向目标服务器发送请求，并收集pprof数据
// 收集到的pprof数据会存储在本地，并根据使用者的请求返回给使用者
// 结果的显示渲染成html返回到浏览器
type AutoPprofServer struct {
	port     int
	storage  *Storage        // 缓存所有拉取的pprof对象
	config   *config.Manager // 添加配置管理器
	allFetch sync.Map        // 所有正在采集的地址
}

func NewAutoPprofServer(cfgPath string) *AutoPprofServer {
	configManager := config.NewManager(cfgPath)

	// 获取配置文件中的port
	port := configManager.GetConfig().GetPort()

	server := &AutoPprofServer{
		port:    port,
		storage: NewStorage(),
		config:  configManager,
	}

	// 初始化storage
	if err := server.storage.Init(); err != nil {
		slog.Error("init storage failed", "error", err)
		os.Exit(1)
	}

	// 注册配置变更回调
	configManager.OnConfigChange(func(cfg *config.Config) {
		slog.Info("config changed, services count", "count", len(cfg.Services))
	})

	return server
}

func (s *AutoPprofServer) Start() {
	http.Handle("/", http.HandlerFunc(s.handleRequest))
	http.Handle("/fetch/", http.HandlerFunc(s.handleFetchRequest))
	http.Handle("/render/", http.HandlerFunc(s.handleRenderRequest))
	http.Handle("/delete/", http.HandlerFunc(s.handleDeleteRequest))

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("shutting down server...")

		// 关闭storage
		if err := s.storage.Close(); err != nil {
			slog.Error("close storage failed", "error", err)
		}

		os.Exit(0)
	}()

	http.ListenAndServe(":"+strconv.Itoa(s.port), nil)
}

// 首页请求
func (s *AutoPprofServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// 从配置中获取服务列表
	cfg := s.config.GetConfig()
	services := cfg.GetServices()

	// 获取所有的pprof数据
	pprofData := s.storage.GetAllPprof()
	// 构造fetch记录
	fetch := make([]FectRecord, 0)
	for id, p := range pprofData {
		fetch = append(fetch, FectRecord{
			Id:       id,
			Service:  p.Meta.Service,
			Addr:     p.Meta.Addr,
			Type:     p.Meta.Type,
			Seconds:  p.Meta.Seconds,
			Ts:       p.Meta.Ts,
			SavePath: p.Meta.SavePath,
			Status:   p.Meta.Status,
		})
	}

	// 按照时间排序
	sort.Slice(fetch, func(i, j int) bool {
		return fetch[i].Ts > fetch[j].Ts
	})

	// 构造模板数据
	data := struct {
		Services []config.ServiceConfig
		Records  []FectRecord
	}{
		Services: services,
		Records:  fetch,
	}

	// 从嵌入的文件系统中读取模板
	tmpl, err := template.ParseFS(content, "html/main.html")
	if err != nil {
		slog.Error("failed to parse template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		slog.Error("failed to execute template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// 判断是否正在采集
func (s *AutoPprofServer) inFetch(addr string) bool {
	_, loaded := s.allFetch.LoadOrStore(addr, struct{}{})
	return loaded
}

// 释放采集
func (s *AutoPprofServer) releaseFetch(addr string) {
	s.allFetch.Delete(addr)
}

// 获取pprof数据请求
func (s *AutoPprofServer) handleFetchRequest(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	var request RequestFetch
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("failed to decode request", "error", err)
		ResponseFetchError(w, "failed to decode request")
		return
	}
	slog.Debug("fetch pprof", "request", request)

	// 验证url是否在配置里
	cfg := s.config.GetConfig()
	services := cfg.GetServices()
	foundService := config.ServiceConfig{}
	for _, service := range services {
		if service.Addr == request.Addr {
			foundService = service
			break
		}
	}

	// 如果url不在配置里，则返回错误
	if foundService.Addr == "" {
		ResponseFetchError(w, "URL not found in config")
		return
	}

	// 如果正在采集，则返回错误
	if s.inFetch(request.Addr) {
		ResponseFetchError(w, "URL is being profiled")
		return
	}
	defer s.releaseFetch(request.Addr)

	// 根据类型拼成url
	pprofUrl := s.getPprofUrl(foundService, request)
	if pprofUrl == "" {
		ResponseFetchError(w, "URL not found")
		return
	}

	// 拉取pprof数据
	p, err := bridge.FetchPprof(pprofUrl)
	if err != nil {
		slog.Error("failed to fetch pprof", "error", err)
		ResponseFetchError(w, "failed to fetch pprof")
		return
	}
	slog.Debug("fetch pprof success", "filePath", p.SavePath)

	// 保存pprof数据到storage
	id := s.storage.SavePprof(NewProfile(p, ProfileMeta{
		Service:  foundService.Name,
		Addr:     request.Addr,
		Type:     request.Type,
		Seconds:  request.Seconds,
		Ts:       time.Now().Unix(),
		SavePath: p.SavePath,
		Status:   "ok",
	}))

	// 返回文件路径
	ResponseFetchSuccess(w, FectRecord{
		Id:       id,
		Service:  foundService.Name,
		Addr:     request.Addr,
		Type:     request.Type,
		Seconds:  request.Seconds,
		Ts:       time.Now().Unix(),
		SavePath: p.SavePath,
	})
}

// 根据类型拼成url
func (s *AutoPprofServer) getPprofUrl(service config.ServiceConfig, request RequestFetch) string {
	// 如果是cpu，则返回cpu的url， 如果seconds是0，默认30s
	if request.Type == "cpu" {
		if request.Seconds == 0 {
			request.Seconds = 30
		}
		return fmt.Sprintf("http://%s/debug/pprof/profile?seconds=%d", service.Addr, request.Seconds)
	}
	if request.Type == "heap" {
		return fmt.Sprintf("http://%s/debug/pprof/heap", service.Addr)
	}
	return ""
}

// 渲染pprof数据到html的请求
func (s *AutoPprofServer) handleRenderRequest(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	id := r.URL.Query().Get("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ResponseRenderError(w, "invalid id parameter")
		return
	}
	if idInt == 0 {
		ResponseRenderError(w, "id parameter is required")
		return
	}

	// 从storage中获取pprof数据
	p := s.storage.GetPprof(idInt)
	if p == nil {
		ResponseRenderError(w, "pprof not found")
		return
	}
	// 渲染pprof数据到html
	bridge.RenderPprofData(p.Profile, w, r)
	ResponseRenderSuccess(w)
	slog.Debug("render pprof success", "id", id)
}

// 删除pprof数据请求
func (s *AutoPprofServer) handleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	id := r.URL.Query().Get("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ResponseDeleteError(w, "invalid id parameter")
		return
	}
	if idInt == 0 {
		ResponseDeleteError(w, "id parameter is required")
		return
	}

	// 从storage中删除pprof数据
	err = s.storage.DeletePprof(idInt)
	if err != nil {
		ResponseDeleteError(w, err.Error())
		return
	}
	ResponseDeleteSuccess(w, idInt)
}
