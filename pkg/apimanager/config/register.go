package config

import (
	"github.com/gin-gonic/gin"
	//"github.com/gostship/kunkka/pkg/simple/client/s3"
	//"github.com/gostship/kunkka/pkg/simple/client/servicemesh"
	//"github.com/gostship/kunkka/pkg/simple/client/sonarqube"
)

type KunkkaConfig struct {
	//DevopsOptions         *jenkins.Options                   `json:"devops,omitempty" yaml:"devops,omitempty" mapstructure:"devops"`
	//SonarQubeOptions      *sonarqube.Options                 `json:"sonarqube,omitempty" yaml:"sonarQube,omitempty" mapstructure:"sonarqube"`
	//KubernetesOptions  *k8s.KubernetesOptions `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	//ServiceMeshOptions *servicemesh.Options   `json:"servicemesh,omitempty" yaml:"servicemesh,omitempty" mapstructure:"servicemesh"`
	//NetworkOptions     *network.Options       `json:"network,omitempty" yaml:"network,omitempty" mapstructure:"network"`
	//LdapOptions        *ldap.Options          `json:"-,omitempty" yaml:"ldap,omitempty" mapstructure:"ldap"`
	//RedisOptions       *cache.Options         `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	//S3Options             *s3.Options                        `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
	//OpenPitrixOptions     *openpitrix.Options                `json:"openpitrix,omitempty" yaml:"openpitrix,omitempty" mapstructure:"openpitrix"`
	//MonitoringOptions     *prometheus.Options                `json:"monitoring,omitempty" yaml:"monitoring,omitempty" mapstructure:"monitoring"`
	//LoggingOptions        *elasticsearch.Options             `json:"logging,omitempty" yaml:"logging,omitempty" mapstructure:"logging"`
	//AuthenticationOptions *authoptions.AuthenticationOptions `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	//MultiClusterOptions   *multicluster.Options              `json:"multicluster,omitempty" yaml:"multicluster,omitempty" mapstructure:"multicluster"`
	//EventsOptions         *eventsclient.Options              `json:"events,omitempty" yaml:"events,omitempty" mapstructure:"events"`
	//AuditingOptions       *auditingclient.Options            `json:"auditing,omitempty" yaml:"auditing,omitempty" mapstructure:"auditing"`
	//AlertingOptions     *alerting.Options     `json:"alerting,omitempty" yaml:"alerting,omitempty" mapstructure:"alerting"`
	//NotificationOptions *notification.Options `json:"notification,omitempty" yaml:"notification,omitempty" mapstructure:"notification"`
	//AuthorizationOptions  *authorizationoptions.AuthorizationOptions `json:"authorization,omitempty" yaml:"authorization,omitempty" mapstructure:"authorization"`
}

//func NewConf() *KunkkaConfig {
//	return &KunkkaConfig{
//		//DevopsOptions:      jenkins.NewDevopsOptions(),
//		//SonarQubeOptions:   sonarqube.NewSonarQubeOptions(),
//		KubernetesOptions: k8s.NewKubernetesOptions(),
//		//ServiceMeshOptions: servicemesh.NewServiceMeshOptions(),
//		NetworkOptions: network.NewNetworkOptions(),
//		LdapOptions:    ldap.NewOptions(),
//		RedisOptions:   cache.NewRedisOptions(),
//		//S3Options:             s3.NewS3Options(),
//		OpenPitrixOptions:     openpitrix.NewOptions(),
//		MonitoringOptions:     prometheus.NewPrometheusOptions(),
//		AlertingOptions:       alerting.NewAlertingOptions(),
//		NotificationOptions:   notification.NewNotificationOptions(),
//		LoggingOptions:        elasticsearch.NewElasticSearchOptions(),
//		AuthenticationOptions: authoptions.NewAuthenticateOptions(),
//		//AuthorizationOptions:  authorizationoptions.NewAuthorizationOptions(),
//		MultiClusterOptions: multicluster.NewOptions(),
//		//EventsOptions:       eventsclient.NewElasticSearchOptions(),
//		//AuditingOptions:     auditingclient.NewElasticSearchOptions(),
//	}
//}

// return kunkka server config map
func (k *KunkkaConfig) GetConfigMap(c *gin.Context) {
	result := make(map[string]bool, 0)

	result["alerting"] = false
	result["auditing"] = false
	result["authentication"] = true
	result["authorization"] = true
	result["devops"] = false
	result["events"] = false
	result["kubernetes"] = true
	result["logging"] = false
	result["monitoring"] = true
	result["multicluster"] = false
	result["network"] = false
	result["notification"] = false
	result["openpitrix"] = false
	result["redis"] = true
	result["s3"] = false
	result["servicemesh"] = false
	result["sonarqube"] = false
	c.IndentedJSON(200, result)
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
//func (conf *KunkkaConfig) ToMap() map[string]bool {
//	conf.stripEmptyOptions()
//	result := make(map[string]bool, 0)
//
//	if conf == nil {
//		return result
//	}
//
//	c := reflect.Indirect(reflect.ValueOf(conf))
//
//	for i := 0; i < c.NumField(); i++ {
//		name := strings.Split(c.Type().Field(i).Tag.Get("json"), ",")[0]
//		if strings.HasPrefix(name, "-") {
//			continue
//		}
//
//		if c.Field(i).IsNil() {
//			result[name] = false
//		} else {
//			result[name] = true
//		}
//	}
//
//	return result
//}
//
//// Remove invalid options before serializing to json or yaml
//func (conf *KunkkaConfig) stripEmptyOptions() {
//
//	if conf.RedisOptions != nil && conf.RedisOptions.Host == "" {
//		conf.RedisOptions = nil
//	}
//
//	//if conf.DevopsOptions != nil && conf.DevopsOptions.Host == "" {
//	//	conf.DevopsOptions = nil
//	//}
//
//	if conf.MonitoringOptions != nil && conf.MonitoringOptions.Endpoint == "" {
//		conf.MonitoringOptions = nil
//	}
//
//	//if conf.SonarQubeOptions != nil && conf.SonarQubeOptions.Host == "" {
//	//	conf.SonarQubeOptions = nil
//	//}
//
//	if conf.LdapOptions != nil && conf.LdapOptions.Host == "" {
//		conf.LdapOptions = nil
//	}
//
//	if conf.OpenPitrixOptions != nil && conf.OpenPitrixOptions.IsEmpty() {
//		conf.OpenPitrixOptions = nil
//	}
//
//	if conf.NetworkOptions != nil && conf.NetworkOptions.EnableNetworkPolicy == false {
//		conf.NetworkOptions = nil
//	}
//
//	if conf.ServiceMeshOptions != nil && conf.ServiceMeshOptions.IstioPilotHost == "" &&
//		conf.ServiceMeshOptions.ServicemeshPrometheusHost == "" &&
//		conf.ServiceMeshOptions.JaegerQueryHost == "" {
//		conf.ServiceMeshOptions = nil
//	}
//	//
//	//if conf.S3Options != nil && conf.S3Options.Endpoint == "" {
//	//	conf.S3Options = nil
//	//}
//
//	if conf.AlertingOptions != nil && conf.AlertingOptions.Endpoint == "" {
//		conf.AlertingOptions = nil
//	}
//
//	if conf.LoggingOptions != nil && conf.LoggingOptions.Host == "" {
//		conf.LoggingOptions = nil
//	}
//
//	if conf.NotificationOptions != nil && conf.NotificationOptions.Endpoint == "" {
//		conf.NotificationOptions = nil
//	}
//
//	if conf.MultiClusterOptions != nil && !conf.MultiClusterOptions.Enable {
//		conf.MultiClusterOptions = nil
//	}
//
//	//if conf.EventsOptions != nil && conf.EventsOptions.Host == "" {
//	//	conf.EventsOptions = nil
//	//}
//	//
//	//if conf.AuditingOptions != nil && conf.AuditingOptions.Host == "" {
//	//	conf.AuditingOptions = nil
//	//}
//}
