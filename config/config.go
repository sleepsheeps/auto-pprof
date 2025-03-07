package config

import (
	"log/slog"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置结构体
type Config struct {
	Port     int             `yaml:"port"`
	Services []ServiceConfig `yaml:"services"`
}

func (c *Config) GetPort() int {
	return c.Port
}

func (c *Config) GetServices() []ServiceConfig {
	return c.Services
}

// ServiceConfig 服务配置结构体
type ServiceConfig struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

type Manager struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
	onChange   []func(*Config)
}

// NewManager 创建配置管理器
func NewManager(configPath string) *Manager {
	m := &Manager{
		configPath: configPath,
		onChange:   make([]func(*Config), 0),
	}

	// 首次加载配置
	if err := m.Load(); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 启动配置自动重载
	go m.watchConfig()

	return m
}

// Load 加载配置文件
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	m.mu.Lock()
	m.config = &config
	m.mu.Unlock()

	// 通知所有监听者
	m.notifyChange()

	return nil
}

// GetConfig 获取当前配置
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// OnConfigChange 注册配置变更回调函数
func (m *Manager) OnConfigChange(f func(*Config)) {
	m.mu.Lock()
	m.onChange = append(m.onChange, f)
	m.mu.Unlock()
}

// notifyChange 通知所有监听者配置已变更
func (m *Manager) notifyChange() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, f := range m.onChange {
		f(m.config)
	}
}

// watchConfig 监控配置文件变化
func (m *Manager) watchConfig() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var lastModTime time.Time
	for range ticker.C {
		stat, err := os.Stat(m.configPath)
		if err != nil {
			slog.Error("failed to stat config file", "error", err)
			continue
		}

		// 如果文件修改时间变化，重新加载配置
		if stat.ModTime() != lastModTime {
			if err := m.Load(); err != nil {
				slog.Error("failed to reload config", "error", err)
				continue
			}
			lastModTime = stat.ModTime()
			slog.Info("config reloaded successfully")
		}
	}
}
