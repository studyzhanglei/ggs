/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package servicecomb

import (
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/router"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/go-chassis/go-archaius/event"

	"strings"

	"github.com/leon-yc/ggs/internal/core/common"
	wp "github.com/leon-yc/ggs/internal/core/router/weightpool"
)

type routeRuleEventListener struct{}

// update route rule of a service
func (r *routeRuleEventListener) Event(e *event.Event) {
	if e == nil {
		qlog.Warn("Event pointer is nil")
		return
	}
	qlog.WithFields(qlog.Fields{
		"key":   e.Key,
		"event": e.EventType,
		"rule":  e.Value,
	}).Info("dark launch event")
	var service string
	var isV2 bool
	if strings.HasPrefix(e.Key, DarkLaunchPrefix) {
		service = strings.Replace(e.Key, DarkLaunchPrefix, "", 1)
	}
	if strings.HasPrefix(e.Key, DarkLaunchPrefixV2) {
		service = strings.Replace(e.Key, DarkLaunchPrefixV2, "", 1)
		isV2 = true
	}
	raw, ok := e.Value.(string)
	if !ok {
		qlog.WithField("value", raw).Error("invalid route rule")
	}
	switch e.EventType {
	case common.Update:
		SaveRouteRule(service, raw, isV2)
	case common.Create:
		SaveRouteRule(service, raw, isV2)
	case common.Delete:
		cseRouter.DeleteRouteRuleByKey(service)
		qlog.WithField("key", e.Key).Info("route rule is removed")
	}

}

//SaveRouteRule save event rule to local cache
func SaveRouteRule(service string, raw string, isV2 bool) {
	var routeRules []*config.RouteRule
	var err error
	if !isV2 {
		routeRules, err = ConvertJSON2RouteRule(raw)
		if err != nil {
			qlog.WithError(err).Error("LoadRules route rule failed")
		}
	} else {
		originRule, err := config.NewServiceRule(raw)
		if err != nil {
			qlog.WithError(err).Error("LoadRules route rule failed")
			return
		}
		routeRules = originRule.Value()
	}

	validateAndUpdate(routeRules, service)
}

func validateAndUpdate(routeRules []*config.RouteRule, service string) {
	if router.ValidateRule(map[string][]*config.RouteRule{service: routeRules}) {
		cseRouter.SetRouteRuleByKey(service, routeRules)
		wp.GetPool().Reset(service)
		qlog.WithFields(qlog.Fields{
			"service": service,
			"rule":    routeRules,
		}).Info("update route rule success")
	}
}
