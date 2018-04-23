package smg

import (
	"testing"

	"github.com/geniuscirno/smg/configurator"
	"github.com/stretchr/testify/assert"
)

type scenarioParseConfiguratorTarget struct {
	target       string
	expectTarget configurator.Target
	expectErr    error
}

func (s *scenarioParseConfiguratorTarget) test(t *testing.T) {
	assert := assert.New(t)

	target, err := parseConfiguratorTarget(s.target)

	assert.Equal(s.expectErr, err)
	assert.Equal(s.expectTarget, target)
}

var testScenarioParseConfiguratorTarget = []scenarioParseConfiguratorTarget{
	{
		target: "etcd://10.10.10.30:2379/test",
		expectTarget: configurator.Target{
			Scheme:    "etcd",
			Authority: "10.10.10.30:2379",
			Endpoint:  "test",
		},
	},
	{
		target: "etcd://10.10.10.30:2379/",
		expectTarget: configurator.Target{
			Scheme:    "etcd",
			Authority: "10.10.10.30:2379",
			Endpoint:  "",
		},
	},
	{
		target: "etcd://10.10.10.30:2379/test/test1",
		expectTarget: configurator.Target{
			Scheme:    "etcd",
			Authority: "10.10.10.30:2379",
			Endpoint:  "test/test1",
		},
	},
}

func Test_parseConfiguratorTarget(t *testing.T) {
	for _, s := range testScenarioParseConfiguratorTarget {
		s.test(t)
	}
}
