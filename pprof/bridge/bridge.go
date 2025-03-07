package bridge

import (
	"log/slog"
	"net/http"

	"github.com/google/pprof/internal/driver"
	"github.com/google/pprof/profile"
)

// FetchPprof 从url拉取pprof数据，并存到本地, 需要返回文件的路径
func FetchPprof(url string) (*profile.Profile, error) {
	p, err := driver.FetchPprof(url)
	if err != nil {
		slog.Error("fetch pprof failed", "error", err)
		return nil, err
	}
	return p, nil
}

// LoadPprofData 从本地加载pprof数据
func LoadPprofData(filePath string) (*profile.Profile, error) {
	p, err := driver.LoadPprofData(filePath)
	if err != nil {
		slog.Error("load pprof data failed", "error", err)
		return nil, err
	}
	return p, nil
}

// RenderPprofData 渲染pprof数据到html
func RenderPprofData(p *profile.Profile, w http.ResponseWriter, r *http.Request) (err error) {
	if err = driver.RenderPprofData(p, w, r); err != nil {
		slog.Error("render pprof data failed", "error", err)
		return err
	}
	return nil
}
