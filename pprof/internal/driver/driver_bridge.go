package driver

import (
	"bufio"
	"net/http"
	"os"

	"github.com/google/pprof/internal/binutils"
	"github.com/google/pprof/internal/plugin"
	"github.com/google/pprof/internal/symbolizer"
	"github.com/google/pprof/profile"
)

// 桥接pprof库的driver包
func FetchPprof(url string) (*profile.Profile, error) {
	src := &source{
		Sources: []string{url},
	}
	p, err := fetchProfiles(src, setDefaultsOverride(nil))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// LoadPprofData 加载pprof数据到内存中
func LoadPprofData(filePath string) (*profile.Profile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	p, err := profile.Parse(file)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// 渲染pprof数据到html
func RenderPprofData(p *profile.Profile, w http.ResponseWriter, r *http.Request) error {
	copier := makeProfileCopier(p)
	ui, err := makeWebInterface(p, copier, setDefaultsOverride(nil))
	if err != nil {
		return err
	}
	ui.stackView(w, r)
	return nil
}

func setDefaultsOverride(o *plugin.Options) *plugin.Options {
	d := &plugin.Options{}
	if o != nil {
		*d = *o
	}
	if d.Writer == nil {
		d.Writer = oswriter{}
	}
	if d.Flagset == nil {
		d.Flagset = &GoFlags{}
	}
	if d.Obj == nil {
		d.Obj = &binutils.Binutils{}
	}
	if d.UI == nil {
		d.UI = &stdUI{r: bufio.NewReader(os.Stdin)}
	}
	if d.HTTPTransport == nil {
		// d.HTTPTransport = transport.New(d.Flagset)
	}
	if d.Sym == nil {
		d.Sym = &symbolizer.Symbolizer{Obj: d.Obj, UI: d.UI, Transport: d.HTTPTransport}
	}
	return d
}
