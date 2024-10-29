package main

import (
	"gladiatorsGoModule/k8s"
	"gladiatorsGoModule/setting"
	"lobby/logger"

	log "github.com/sirupsen/logrus"
)

// getExternalIP 取Loadbalancer分配給此pod的對外IP
func getExternalIP() (string, error) {
	ip, err := k8s.GetLoadBalancerExternalIP(setting.NAMESPACE_LOBBY, setting.LOBBY)
	if err != nil {
		log.Errorf("%s GetLoadBalancerExternalIP error: %v.\n", logger.LOG_Main, err)
	}
	return ip, err
}
