package metricsink

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chassis/go-chassis/v2/core/common"
	"github.com/go-chassis/go-chassis/v2/core/config"
	"github.com/go-chassis/go-chassis/v2/core/endpoint"
	chassisTLS "github.com/go-chassis/go-chassis/v2/core/tls"
	"github.com/go-chassis/openlog"
)

// constants for header parameters
const (
	//HeaderUserName is a variable of type string
	HeaderUserName   = "x-user-name"
	HeaderDomainName = "x-domain-name"
	ContentType      = "Content-Type"
	Name             = "monitor"
)

func getTLSForClient(monitorURL string) (*tls.Config, error) {
	monitorServerURL, err := url.Parse(monitorURL)
	if err != nil {
		openlog.GetLogger().Error("Error occurred while parsing Monitor Server Uri" + err.Error())
		return nil, err
	}
	scheme := monitorServerURL.Scheme
	if scheme != "https" {
		return nil, nil
	}

	sslTag := Name + "." + common.Consumer
	tlsConfig, sslConfig, err := chassisTLS.GetTLSConfigByService(Name, "", common.Consumer)
	if err != nil {
		if chassisTLS.IsSSLConfigNotExist(err) {
			return nil, fmt.Errorf("%s TLS mode, but no ssl config", sslTag)
		}
		return nil, err
	}
	openlog.Warn(fmt.Sprintf("%s TLS mode, verify peer: %t, cipher plugin: %s",
		sslTag, sslConfig.VerifyPeer, sslConfig.CipherPlugin))

	return tlsConfig, nil
}

func getAuthHeaders() http.Header {
	userName := config.GlobalDefinition.ServiceComb.Monitor.Client.UserName
	if userName == "" {
		userName = common.DefaultUserName
	}
	domainName := config.GlobalDefinition.ServiceComb.Monitor.Client.DomainName
	if domainName == "" {
		domainName = common.DefaultDomainName
	}

	headers := make(http.Header)
	headers.Set(HeaderUserName, userName)
	headers.Set(HeaderDomainName, domainName)
	headers.Set(ContentType, "application/json")

	return headers
}

func getMonitorEndpoint() (string, error) {
	monitorEndpoint := config.GlobalDefinition.ServiceComb.Monitor.Client.ServerURI
	if monitorEndpoint == "" {
		monitorURL, err := endpoint.GetEndpoint("default", "CseMonitoring", "latest")
		if err != nil {
			openlog.Warn(fmt.Sprintf("empty monitor server endpoint, please provide the monitor server endpoint, err: %v", err))
			return "", err
		}

		monitorEndpoint = monitorURL
	}

	return monitorEndpoint, nil
}
