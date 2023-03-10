package services

import (
	"isula.org/rubik/pkg/feature"
	"isula.org/rubik/pkg/services/dynCache"
	"isula.org/rubik/pkg/services/helper"
	"isula.org/rubik/pkg/services/iocost"
	"isula.org/rubik/pkg/services/iolimit"
	"isula.org/rubik/pkg/services/preemption"
	"isula.org/rubik/pkg/services/quotaburst"
	"isula.org/rubik/pkg/services/quotaturbo"
)

type ServiceComponent func(name string) error

var (
	serviceComponents = map[string]ServiceComponent{
		feature.FeaturePreemption: initPreemptionFactory,
		feature.FeatureDynCache:   initDynCacheFactory,
		feature.FeatureIOLimit:    initIOLimitFactory,
		feature.FeatureIOCost:     initIOCostFactory,
		feature.FeatureDynMemory:  initDynCacheFactory,
		feature.FeatureQuotaBurst: initQuotaBurstFactory,
		feature.FeatureQuotaTurbo: initQuotaTurboFactory,
	}
)

func initIOLimitFactory(name string) error {
	helper.AddFactory(name, iolimit.IOLimitFactory{ObjName: name})
	return nil
}

func initIOCostFactory(name string) error {
	helper.AddFactory(name, iocost.IOCostFactory{ObjName: name})
	return nil
}

func initDynCacheFactory(name string) error {
	helper.AddFactory(name, dynCache.DynCacheFactory{ObjName: name})
	return nil
}

func initQuotaTurboFactory(name string) error {
	helper.AddFactory(name, quotaturbo.QuotaTurboFactory{ObjName: name})
	return nil
}

func initQuotaBurstFactory(name string) error {
	helper.AddFactory(name, quotaburst.BurstFactory{ObjName: name})
	return nil
}

func initPreemptionFactory(name string) error {
	helper.AddFactory(name, preemption.PreemptionFactory{ObjName: name})
	return nil
}
