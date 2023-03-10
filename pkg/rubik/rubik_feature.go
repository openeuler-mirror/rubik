package rubik

import (
	"isula.org/rubik/pkg/feature"
	"isula.org/rubik/pkg/services"
)

var defaultRubikFeature = []services.FeatureSpec{
	{
		Name:    feature.FeaturePreemption,
		Default: true,
	},
	{
		Name:    feature.FeatureDynCache,
		Default: true,
	},
	{
		Name:    feature.FeatureIOLimit,
		Default: true,
	},
	{
		Name:    feature.FeatureIOCost,
		Default: true,
	},
	{
		Name:    feature.FeatureDynMemory,
		Default: true,
	},
	{
		Name:    feature.FeatureQuotaBurst,
		Default: true,
	},
	{
		Name:    feature.FeatureQuotaTurbo,
		Default: true,
	},
}
