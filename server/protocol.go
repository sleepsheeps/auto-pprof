package server

import (
	"encoding/json"
	"net/http"
)

// 拉取pprof数据请求
type RequestFetch struct {
	Addr    string `json:"addr"`
	Type    string `json:"type"`    // cpu mem
	Seconds int    `json:"seconds"` // 采集时间
}

// 拉取pprof数据响应
type ResponseFetch struct {
	Error string     `json:"error"`
	Fetch FectRecord `json:"fetch"`
}

// 拉取pprof数据记录
type FectRecord struct {
	Id       int64  `json:"id"`
	Service  string `json:"service"`
	Addr     string `json:"addr"`
	Type     string `json:"type"`      // cpu mem
	Seconds  int    `json:"seconds"`   // 采集时间
	Ts       int64  `json:"ts"`        // 采集时间戳
	SavePath string `json:"save_path"` // 保存路径
	Status   string `json:"status"`    // ok, error
}

// 拉取pprof数据成功
func ResponseFetchSuccess(w http.ResponseWriter, fetch FectRecord) {
	response := ResponseFetch{
		Error: "",
		Fetch: fetch,
	}
	json.NewEncoder(w).Encode(response)
}

// 拉取pprof数据失败
func ResponseFetchError(w http.ResponseWriter, error string) {
	response := ResponseFetch{
		Error: error,
	}
	json.NewEncoder(w).Encode(response)
}

// 渲染pprof数据到html的响应
type ResponseRender struct {
	Error string `json:"error"`
}

// 渲染pprof数据到html成功
func ResponseRenderSuccess(w http.ResponseWriter) {
	response := ResponseRender{
		Error: "",
	}
	json.NewEncoder(w).Encode(response)
}

// 渲染pprof数据到html失败
func ResponseRenderError(w http.ResponseWriter, error string) {
	response := ResponseRender{
		Error: error,
	}
	json.NewEncoder(w).Encode(response)
}

// 删除接口
type ResponseDelete struct {
	Id    int64  `json:"id"`
	Error string `json:"error"`
}

// 删除失败
func ResponseDeleteError(w http.ResponseWriter, error string) {
	response := ResponseDelete{
		Id:    -1,
		Error: error,
	}
	json.NewEncoder(w).Encode(response)
}

// 删除成功
func ResponseDeleteSuccess(w http.ResponseWriter, id int64) {
	response := ResponseDelete{
		Id:    id,
		Error: "",
	}
	json.NewEncoder(w).Encode(response)
}
