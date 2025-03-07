package server

import (
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/google/pprof/bridge"
	"github.com/syndtr/goleveldb/leveldb"
)

// 缓存所有拉取的pprof对象, 这样处理请求直接在此模块去获取pprof对象
// 如果pprof对象不存在，则从文件中去读取
type Storage struct {
	id    int64
	m     sync.Map
	count int64
	db    *leveldb.DB // leveldb实例
}

func NewStorage() *Storage {
	return &Storage{
		id:    0,
		count: 0,
	}
}

func (s *Storage) Init() error {
	// 打开leveldb
	db, err := leveldb.OpenFile("saved", nil)
	if err != nil {
		return err
	}
	s.db = db

	// 从leveldb中读取所有meta数据
	iter := s.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		// 反序列化meta数据
		var meta ProfileMeta
		if err := json.Unmarshal(value, &meta); err != nil {
			slog.Error("unmarshal meta failed", "error", err)
			continue
		}

		// 从key中解析id
		var id int64
		if err := json.Unmarshal(key, &id); err != nil {
			slog.Error("unmarshal id failed", "error", err)
			continue
		}

		// 更新最大id
		if id > s.id {
			s.id = id
		}

		// 构造Profile对象并保存到内存
		profile := &Profile{
			Meta: meta,
		}
		s.m.Store(id, profile)
		atomic.AddInt64(&s.count, 1)
	}
	iter.Release()

	slog.Info("init storage from leveldb", "count", s.count)
	return nil
}

func (s *Storage) GetPprof(id int64) *Profile {
	p, ok := s.m.Load(id)
	if !ok {
		slog.Error("pprof not found", "id", id)
		return nil
	}
	loadProfile := p.(*Profile)
	// 如果是空则需要从文件中解析出来
	if loadProfile.Profile == nil {
		// 从文件中读取pprof数据
		profile, err := bridge.LoadPprofData(loadProfile.Meta.SavePath)
		if err != nil {
			loadProfile.Meta.Status = "error"
			slog.Error("load pprof data failed", "error", err)
			return nil
		}
		loadProfile.Profile = profile
	}
	return loadProfile
}

func (s *Storage) SavePprof(p *Profile) int64 {
	// 自增id
	id := atomic.AddInt64(&s.id, 1)

	// 保存meta数据到leveldb
	key, _ := json.Marshal(id)
	value, _ := json.Marshal(p.Meta)
	if err := s.db.Put(key, value, nil); err != nil {
		slog.Error("save meta to leveldb failed", "error", err)
	}

	// 保存到内存
	s.m.Store(id, p)
	atomic.AddInt64(&s.count, 1)
	slog.Debug("save pprof", "id", id, "meta", p.Meta)
	return id
}

func (s *Storage) GetAllPprof() map[int64]*Profile {
	pprofMap := make(map[int64]*Profile)
	s.m.Range(func(key, value interface{}) bool {
		pprofMap[key.(int64)] = value.(*Profile)
		return true
	})
	slog.Debug("get all pprof", "count", s.count)
	return pprofMap
}

// 删除pprof数据
func (s *Storage) DeletePprof(id int64) error {
	v, _ := s.m.LoadAndDelete(id)
	if v != nil {
		// 从leveldb中删除
		key, _ := json.Marshal(id)
		if err := s.db.Delete(key, nil); err != nil {
			slog.Error("delete meta from leveldb failed", "error", err)
		}

		atomic.AddInt64(&s.count, -1)
		slog.Info("delete pprof", "id", id, "pprof", v)
	}
	return nil
}

// 在程序退出时调用Close
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
