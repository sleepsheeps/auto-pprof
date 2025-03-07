package server

import "github.com/google/pprof/profile"

type Profile struct {
	*profile.Profile
	Meta ProfileMeta
}

// ProfileMeta 拉取pprof数据时，返回的元数据
type ProfileMeta struct {
	Service  string `json:"service"`
	Addr     string `json:"addr"`
	Type     string `json:"type"`
	Seconds  int    `json:"seconds"`
	Ts       int64  `json:"ts"`
	SavePath string `json:"save_path"`
	Status   string `json:"status"` // ok, error
}

func NewProfile(p *profile.Profile, meta ProfileMeta) *Profile {
	return &Profile{
		Profile: p,
		Meta:    meta,
	}
}
