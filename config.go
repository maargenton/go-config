package config

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/copier"
	"github.com/marcus999/go-config/pkg/debounce"
	"github.com/marcus999/go-config/pkg/watch"
)

// Loader loads and watches config
type Loader struct {
	filename      string
	defaultConfig interface{}
	config        atomic.Value
	watcher       *watch.FileWatcher

	reloadHandlers     []func(interface{})
	errorHandlers      []func(error)
	validationHandlers []func(interface{}) (interface{}, error)
	strictParsing      bool
	keepLastValid      bool
	debounceInterval   time.Duration
	debounceMaxDelay   time.Duration
}

// Option is the base tupe for configuration options
type Option func(*Loader)

const (
	// DefaultDebounceInterval defines the default debounce interval of 100ms
	DefaultDebounceInterval = 1000 * time.Millisecond

	// DefaultDebounceMaxDelay defines the default debounce max delay of 3s
	DefaultDebounceMaxDelay = 3 * time.Second
)

// ---------------------------------------------------------------------------
// config loader options
// ---------------------------------------------------------------------------

// ReloadHandler attaches a function to be called when the configuration is
// reloaded
func ReloadHandler(f func(interface{})) Option {
	return func(c *Loader) {
		c.reloadHandlers = append(c.reloadHandlers, f)
	}
}

// ErrorHandler attaches a function to be called when an error occurs during
// a background opration, e.g. while reloading the configuration file
func ErrorHandler(f func(err error)) Option {
	return func(c *Loader) {
		c.errorHandlers = append(c.errorHandlers, f)
	}
}

// ValidationHandler attaches a function to be called when a new configuration
// is loaded, but before it propagates through the system. The handler can
// modify or enhance the config object, and can abort the update by returning
// an error.
func ValidationHandler(f func(interface{}) (interface{}, error)) Option {
	return func(c *Loader) {
		c.validationHandlers = append(c.validationHandlers, f)
	}
}

// OptStrictParsing activate the strict option for the underlying parsing of
// the configuration file, i.e. fields that are unknown generate an error
// rather than being silently ignored
func OptStrictParsing() Option {
	return func(c *Loader) {
		c.strictParsing = true
	}
}

// OptKeepLatestOnFailure activate an option that will keep the latest valid
// configuration if the new configuration fails to load. The default behavior
// is to revert to the default settings. This option is not recommended for
// application running in a cluster orchestration environment.
func OptKeepLatestOnFailure() Option {
	return func(c *Loader) {
		c.keepLastValid = true
	}
}

// OptDebounceInterval set the debounce interval for rapid changes to the
// configuration file. Default interval is 100ms
func OptDebounceInterval(v time.Duration) Option {
	return func(c *Loader) {
		c.debounceInterval = v
	}
}

// OptDebounceMaxDelay set the maximum delay that can be caused by the
// debouncing process, in case the target configuration file is constantly
// cahning. The default value is 3s.
func OptDebounceMaxDelay(v time.Duration) Option {
	return func(c *Loader) {
		c.debounceMaxDelay = v
	}
}

// ---------------------------------------------------------------------------
// config loader interface
// ---------------------------------------------------------------------------

// NewLoader creates a new configuration loader from a filename and a set of defaults
func NewLoader(filename string, defaultConfig interface{}, opts ...Option) (*Loader, error) {

	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	w, err := watch.NewFileWatcher(filename)
	if err != nil {
		return nil, err
	}

	c := &Loader{
		filename:         filename,
		defaultConfig:    normalizeToSinglePtr(defaultConfig),
		watcher:          w,
		debounceInterval: DefaultDebounceInterval,
		debounceMaxDelay: DefaultDebounceInterval,
	}

	for _, opt := range opts {
		opt(c)
	}

	cfg := cloneStruct(c.defaultConfig)
	err = c.loadConfigFile(filename, cfg)
	if err != nil {
		c.handleError(err)
	}

	c.applyValidations(cfg)
	c.config.Store(cfg)

	if c.debounceInterval != 0 {
		in, out := debounce.New(c.debounceInterval, c.debounceMaxDelay)
		go func() {
			for {
				e, ok := <-c.watcher.UpdateChannel()
				if !ok {
					return
				}
				log.Printf("watcher event: %v", e)
				in <- debounce.Event
			}
		}()
		go func() {
			for {
				_, ok := <-out
				if !ok {
					return
				}
				log.Printf("debounce event\n")
				c.reloadConfig()
			}
		}()
	} else {
		go func() {
			for {
				_, ok := <-c.watcher.UpdateChannel()
				if !ok {
					return
				}
				c.reloadConfig()
			}
		}()
	}

	return c, nil
}

// Error types:
//  - Abs : fatal error
// Failed to load config - using defaults
// Failed to watch config - n/a

// Get returns the current version of the configuraiton stored in the loader
func (c *Loader) Get() interface{} {
	return c.config.Load()
}

// GetDefaults returns a copy of the default config
func (c *Loader) GetDefaults() interface{} {
	return c.defaultConfig
}

// ---------------------------------------------------------------------------
// config loader implemetation
// ---------------------------------------------------------------------------

func (c *Loader) loadConfigFile(filename string, cfg interface{}) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var opts []yaml.JSONOpt
	if c.strictParsing {
		opts = append(opts, yaml.DisallowUnknownFields)
	}

	err = yaml.Unmarshal(content, cfg, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Loader) reloadConfig() {
	cfg := cloneStruct(c.defaultConfig)
	err := c.loadConfigFile(c.filename, cfg)
	if err != nil {
		c.handleError(err)
		if c.keepLastValid {
			return
		} else {
			cfg = cloneStruct(c.defaultConfig)
		}
	}

	c.applyValidations(cfg)
	c.config.Store(cfg)
	c.notifyReloadHandlers(cfg)
}

func (c *Loader) notifyReloadHandlers(cfg interface{}) {
	for _, handler := range c.reloadHandlers {
		handler(cfg)
	}
}

func (c *Loader) handleError(err error) {
	for _, handler := range c.errorHandlers {
		handler(err)
	}
}

func (c *Loader) applyValidations(cfg interface{}) (interface{}, error) {
	for _, validate := range c.validationHandlers {
		var err error
		cfg, err = validate(cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// ---------------------------------------------------------------------------
// configuration struct helpers
// ---------------------------------------------------------------------------

func cloneStruct(v interface{}) interface{} {
	var cfg = reflect.New(reflect.TypeOf(v).Elem()).Interface()
	copier.Copy(cfg, v)
	return cfg
}

func normalizeToSinglePtr(v interface{}) interface{} {
	baseType := reflect.TypeOf(v)
	for baseType.Kind() == reflect.Ptr {
		baseType = baseType.Elem()
	}

	rv := reflect.ValueOf(v)
	rvp := rv
	for rv.Kind() == reflect.Ptr {
		rvp = rv
		rv = rv.Elem()
	}

	if rvp == rv {
		rvp = reflect.New(baseType)
		copier.Copy(rvp.Interface(), rv.Interface())
	}

	return rvp.Interface()
}
