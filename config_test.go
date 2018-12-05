package config_test

import (
	"reflect"
	"testing"

	"github.com/marcus999/go-config"

	"github.com/marcus999/go-testpredicate"
	"github.com/marcus999/go-testpredicate/pred"
)

type testConfig struct {
	Name string
	Port int
}

var testConfigDefaults = testConfig{
	Name: "defaultName",
	Port: 1234,
}

// ---------------------------------------------------------------------------
// Test config defaults
// ---------------------------------------------------------------------------

func TestGetDefaultsWithDefaultAsValue(t *testing.T) {
	assert := testpredicate.NewAsserter(t)

	c, err := config.NewLoader("a/b/c.yaml", testConfigDefaults)
	assert.That(c, pred.IsNotNil())
	assert.That(err, pred.IsNil())

	icfg := c.GetDefaults()
	assert.That(reflect.TypeOf(icfg), pred.IsEqualTo(reflect.TypeOf(&testConfigDefaults)))

	cfg := icfg.(*testConfig)
	assert.That(cfg.Name, pred.IsEqualTo(testConfigDefaults.Name))
	assert.That(cfg.Port, pred.IsEqualTo(testConfigDefaults.Port))
}

func TestGetDefaultsWithDefaultAsPtr(t *testing.T) {
	assert := testpredicate.NewAsserter(t)

	c, err := config.NewLoader("a/b/c.yaml", &testConfigDefaults)
	assert.That(c, pred.IsNotNil())
	assert.That(err, pred.IsNil())

	icfg := c.GetDefaults()
	assert.That(reflect.TypeOf(icfg), pred.IsEqualTo(reflect.TypeOf(&testConfigDefaults)))

	cfg := icfg.(*testConfig)
	assert.That(cfg.Name, pred.IsEqualTo(testConfigDefaults.Name))
	assert.That(cfg.Port, pred.IsEqualTo(testConfigDefaults.Port))
}

func TestGetDefaultsWithDefaultAsPtrPtr(t *testing.T) {
	assert := testpredicate.NewAsserter(t)

	d := &testConfigDefaults
	c, err := config.NewLoader("a/b/c.yaml", &d)
	assert.That(c, pred.IsNotNil())
	assert.That(err, pred.IsNil())

	icfg := c.GetDefaults()
	assert.That(reflect.TypeOf(icfg), pred.IsEqualTo(reflect.TypeOf(&testConfigDefaults)))

	cfg := icfg.(*testConfig)
	assert.That(cfg.Name, pred.IsEqualTo(testConfigDefaults.Name))
	assert.That(cfg.Port, pred.IsEqualTo(testConfigDefaults.Port))
}

// ---------------------------------------------------------------------------
// Test config defaults
// ---------------------------------------------------------------------------

func TestWithDefaultAsPtr(t *testing.T) {
	assert := testpredicate.NewAsserter(t)

	c, err := config.NewLoader("a/b/c.yaml", &testConfigDefaults)
	assert.That(c, pred.IsNotNil())
	assert.That(err, pred.IsNil())

	icfg := c.GetDefaults()
	assert.That(reflect.TypeOf(icfg), pred.IsEqualTo(reflect.TypeOf(&testConfigDefaults)))

	cfg := icfg.(*testConfig)
	assert.That(cfg.Name, pred.IsEqualTo(testConfigDefaults.Name))
	assert.That(cfg.Port, pred.IsEqualTo(testConfigDefaults.Port))
}
