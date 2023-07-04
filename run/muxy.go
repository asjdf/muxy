// Package run contains the main Muxy entrypoints.
package run

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mefellows/muxy/log"
	"github.com/mefellows/muxy/muxy"
	"github.com/mefellows/plugo/plugo"
)

// Config is the top-level configuration struct
type Config struct {
	RawConfig  *plugo.RawConfig
	ConfigFile string // Path to YAML Configuration File
}

// PluginConfig contains configuration for all plugins to Muxy
type PluginConfig struct {
	Name        string
	Description string
	LogLevel    int `default:"2" required:"true" mapstructure:"loglevel"`
	Proxy       []plugo.PluginConfig
	Middleware  []plugo.PluginConfig
}

// Muxy is the main orchestration component
type Muxy struct {
	config      *Config
	middlewares []muxy.Middleware
	proxies     []muxy.Proxy
	sigChan     chan os.Signal
}

// New creates a new Muxy instance
func New(config *Config) *Muxy {
	return &Muxy{config: config}
}

// NewWithDefaultConfig creates a new Muxy instance with defaults
func NewWithDefaultConfig() *Muxy {
	c := &Config{}
	return &Muxy{config: c}
}

// Run the mucking proxy!
func (m *Muxy) Run() {
	m.LoadPlugins()

	// Setup all plugins...
	for _, m := range m.middlewares {
		m.Setup()
	}

	// Interrupt handler
	m.sigChan = make(chan os.Signal, 1)
	signal.Notify(m.sigChan, os.Interrupt, syscall.SIGTERM)

	// Start proxy
	for _, proxy := range m.proxies {
		go proxy.Proxy()
	}

	// Block until a signal is received.
	<-m.sigChan
	log.Info("Shutting down Muxy...")

	for _, m := range m.middlewares {
		m.Teardown()
	}

	for _, p := range m.proxies {
		p.Teardown()
	}
}

// LoadPlugins loads all plugins dynamically and configures them
func (m *Muxy) LoadPlugins() {
	// Load Configuration
	var err error
	var confLoader *plugo.ConfigLoader
	c := &PluginConfig{}
	if m.config.ConfigFile != "" {
		confLoader = &plugo.ConfigLoader{}
		err = confLoader.LoadFromFile(m.config.ConfigFile, &c)
		if err != nil {
			log.Fatalf("Unable to read configuration file: %s", err.Error())
		}
	} else {
		log.Fatal("No config file provided")
	}

	log.SetLevel(log.Level(c.LogLevel))

	// Load all plugins
	m.middlewares = make([]muxy.Middleware, len(c.Middleware))
	plugins := plugo.LoadPluginsWithConfig(confLoader, c.Middleware)
	for i, p := range plugins {
		log.Info("Loading plugin \t" + log.Colorize(log.YELLOW, c.Middleware[i].Name))
		m.middlewares[i] = p.(muxy.Middleware)
	}

	m.proxies = make([]muxy.Proxy, len(c.Proxy))
	plugins = plugo.LoadPluginsWithConfig(confLoader, c.Proxy)
	for i, p := range plugins {
		log.Info("Loading proxy \t" + log.Colorize(log.YELLOW, c.Proxy[i].Name))
		m.proxies[i] = p.(muxy.Proxy)
		m.proxies[i].Setup(m.middlewares)
	}
}
