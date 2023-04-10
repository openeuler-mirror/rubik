package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/feature"
)

var defaultFeature = []FeatureSpec{
	{
		Name:    feature.PreemptionFeature,
		Default: true,
	},
	{
		Name:    feature.DynCacheFeature,
		Default: true,
	},
	{
		Name:    feature.IOLimitFeature,
		Default: true,
	},
	{
		Name:    feature.IOCostFeature,
		Default: true,
	},
	{
		Name:    feature.DynMemoryFeature,
		Default: true,
	},
	{
		Name:    feature.QuotaBurstFeature,
		Default: true,
	},
	{
		Name:    feature.QuotaTurboFeature,
		Default: true,
	},
}

func TestErrorInitServiceComponents(t *testing.T) {
	errFeatures := []FeatureSpec{
		{
			Name:    "testFeature",
			Default: true,
		},
		{
			Name:    feature.QuotaTurboFeature,
			Default: false,
		},
	}

	InitServiceComponents(errFeatures)
	for _, feature := range errFeatures {
		_, err := GetServiceComponent(feature.Name)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "get service failed")
		}
	}
}

func TestInitServiceComponents(t *testing.T) {
	InitServiceComponents(defaultFeature)
	for _, feature := range defaultFeature {
		s, err := GetServiceComponent(feature.Name)
		if err != nil {
			assert.Contains(t, err.Error(), "this machine not support")
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, s.ID(), feature.Name)
	}
}
