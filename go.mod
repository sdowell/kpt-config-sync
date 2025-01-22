module kpt.dev/configsync

go 1.23

require (
	cloud.google.com/go/compute/metadata v0.6.0
	cloud.google.com/go/monitoring v1.22.1
	cloud.google.com/go/trace v1.11.2
	contrib.go.opencensus.io/exporter/ocagent v0.7.0
	github.com/GoogleContainerTools/kpt v1.0.0-beta.46
	github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-namespace v0.4.1-0.20220713210718-d955e7d3a800
	github.com/GoogleContainerTools/kpt-functions-sdk/go/fn v0.0.0-20220706221933-7181f451a663
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/elliotchance/orderedmap/v2 v2.2.0
	github.com/ettle/strcase v0.2.0
	github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/go-logr/logr v1.4.2
	github.com/golang/protobuf v1.5.4
	github.com/google/addlicense v1.1.1
	github.com/google/gnostic-models v0.6.8
	github.com/google/go-cmp v0.6.0
	github.com/google/go-containerregistry v0.16.1
	github.com/google/go-licenses/v2 v2.0.0-alpha.1
	github.com/google/uuid v1.6.0
	github.com/jstemmer/go-junit-report/v2 v2.0.0
	github.com/kylelemons/godebug v1.1.0
	github.com/open-policy-agent/cert-controller v0.10.2-0.20240531181455-2649f121ab97
	github.com/prometheus/client_golang v1.20.5
	github.com/prometheus/common v0.61.0
	github.com/spf13/cobra v1.8.1
	github.com/spyzhov/ajson v0.9.1
	github.com/stretchr/testify v1.10.0
	go.opencensus.io v0.24.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/exp v0.0.0-20241217172543-b2144cdd0a67
	golang.org/x/mod v0.22.0
	google.golang.org/api v0.214.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.31.3
	k8s.io/apiextensions-apiserver v0.31.0
	k8s.io/apimachinery v0.31.3
	k8s.io/cli-runtime v0.30.3
	k8s.io/client-go v0.31.3
	k8s.io/cluster-registry v0.0.6
	k8s.io/code-generator v0.31.0
	k8s.io/klog/v2 v2.130.1
	k8s.io/kube-aggregator v0.30.3
	k8s.io/kube-openapi v0.0.0-20240430033511-f0e62f92d13f
	k8s.io/kubectl v0.30.3
	k8s.io/kubernetes v1.30.3
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8
	sigs.k8s.io/cli-utils v0.37.2
	sigs.k8s.io/controller-runtime v0.19.3
	sigs.k8s.io/controller-runtime/tools/setup-envtest v0.0.0-20231023142458-b9f29826ee83
	sigs.k8s.io/controller-tools v0.15.0
	sigs.k8s.io/kind v0.23.0 // When upgrading, update the node image versions at e2e/nomostest/clusters/kind.go
	sigs.k8s.io/kustomize/api v0.15.0
	sigs.k8s.io/kustomize/kyaml v0.17.1
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1
	sigs.k8s.io/yaml v1.4.0
)

// dependencies for opentelemetry
require (
	github.com/open-telemetry/opentelemetry-collector-contrib/confmap/provider/s3provider v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/confmap/provider/secretsmanagerprovider v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/datadogconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/exceptionsconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/failoverconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/grafanacloudconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/otlpjsonconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/roundrobinconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/servicegraphconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/sumconnector v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/alertmanagerexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/alibabacloudlogserviceexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awscloudwatchlogsexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsemfexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awskinesisexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsxrayexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/azuredataexplorerexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/azuremonitorexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/carbonexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/cassandraexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/clickhouseexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/coralogixexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/datadogexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/datasetexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dorisexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudpubsubexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlemanagedprometheusexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/honeycombmarkerexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/influxdbexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/logicmonitorexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/logzioexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/lokiexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/mezmoexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/opencensusexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/opensearchexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/pulsarexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/rabbitmqexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sentryexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/syslogexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/tencentcloudlogserviceexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/zipkinexporter v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/ackextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/asapauthextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/awsproxy v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/avrologencodingextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/jaegerencodingextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/jsonlogencodingextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/otlpencodingextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/textencodingextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/zipkinencodingextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/googleclientauthextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckv2extension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarderextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/jaegerremotesampling v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/cfgardenobserver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecstaskobserver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/oidcauthextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/remotetapextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/solarwindsapmsettingsextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/dbstorage v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatocumulativeprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/geoipprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbytraceprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/intervalprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/logdedupprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/remotetapprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/routingprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/sumologicprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/activedirectorydsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsfirehosereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awss3receiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsxrayreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureblobreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/bigipreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/chronyreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudflarereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/couchdbreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/datadogreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/expvarreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filestatsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/flinkmetricsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/githubreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudmonitoringreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudspannerreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/influxdbreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/libhoneyreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/lokireceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/memcachedreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/namedpipereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nsxtreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ntpreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/opencensusreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otlpjsonfilereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/podmanreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/pulsarreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/purefareceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/purefbreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/riakreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/skywalkingreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snmpreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snowflakereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/solacereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkenterprisereceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sshcheckreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/udplogreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wavefrontreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/webhookeventreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver v0.117.0
	go.opentelemetry.io/collector/component v0.117.0
	go.opentelemetry.io/collector/confmap v1.23.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.23.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.23.0
	go.opentelemetry.io/collector/confmap/provider/httpprovider v1.23.0
	go.opentelemetry.io/collector/confmap/provider/httpsprovider v1.23.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.23.0
	go.opentelemetry.io/collector/connector v0.117.0
	go.opentelemetry.io/collector/connector/forwardconnector v0.117.0
	go.opentelemetry.io/collector/exporter v0.117.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.117.0
	go.opentelemetry.io/collector/exporter/nopexporter v0.117.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.117.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.117.0
	go.opentelemetry.io/collector/extension v0.117.0
	go.opentelemetry.io/collector/extension/zpagesextension v0.117.0
	go.opentelemetry.io/collector/otelcol v0.117.0
	go.opentelemetry.io/collector/processor v0.117.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.117.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.117.0
	go.opentelemetry.io/collector/receiver v0.117.0
	go.opentelemetry.io/collector/receiver/nopreceiver v0.117.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.117.0
)

// When updating cloud.google.com/go/auth, ensure the auth.IsCredentialsNotFoundError still works.
require cloud.google.com/go/auth v0.13.0

require (
	bitbucket.org/atlassian/go-asap/v2 v2.9.0 // indirect
	cel.dev/expr v0.16.2 // indirect
	cloud.google.com/go v0.117.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.6 // indirect
	cloud.google.com/go/iam v1.2.2 // indirect
	cloud.google.com/go/logging v1.13.0 // indirect
	cloud.google.com/go/longrunning v0.6.2 // indirect
	cloud.google.com/go/pubsub v1.45.3 // indirect
	cloud.google.com/go/spanner v1.73.0 // indirect
	code.cloudfoundry.org/clock v0.0.0-20180518195852-02e53af36e6c // indirect
	code.cloudfoundry.org/garden v0.0.0-20241023020423-a21e43a17f84 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20241007161556-ec30366c7912 // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/lager/v3 v3.11.0 // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/AthenZ/athenz v1.10.39 // indirect
	github.com/Azure/azure-amqp-common-go/v4 v4.2.0 // indirect
	github.com/Azure/azure-event-hubs-go/v3 v3.6.2 // indirect
	github.com/Azure/azure-kusto-go v0.16.1 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.16.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 v5.7.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor v0.11.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 v4.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.5.0 // indirect
	github.com/Azure/azure-storage-queue-go v0.0.0-20230531184854-c06a8eff66fe // indirect
	github.com/Azure/go-amqp v1.3.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/ClickHouse/ch-go v0.61.5 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.30.0 // indirect
	github.com/Code-Hex/go-generics-cache v1.5.1 // indirect
	github.com/DataDog/agent-payload/v5 v5.0.140 // indirect
	github.com/DataDog/datadog-agent/comp/core/config v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/core/flare/builder v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/core/flare/types v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/core/hostname/hostnameinterface v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/core/log/def v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/core/secrets v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/core/telemetry v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/def v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/logs/agent/config v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/otelcol/logsagentpipeline v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/otelcol/logsagentpipeline/logsagentpipelineimpl v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/exporter/logsagentexporter v0.62.0-devel.0.20241213165407-f95df913d2b7 // indirect
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/metricsclient v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/statsprocessor v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/trace/compression/def v0.60.1 // indirect
	github.com/DataDog/datadog-agent/comp/trace/compression/impl-gzip v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/collector/check/defaults v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/env v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/mock v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/model v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/nodetreemodel v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/setup v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/structure v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/teeconfig v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/utils v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/auditor v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/client v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/diagnostic v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/message v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/metrics v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/pipeline v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/processor v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/sds v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/sender v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/sources v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/status/statusinterface v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/logs/status/utils v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/proto v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/status/health v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/telemetry v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/trace v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/backoff v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/cgroups v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/executable v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/filesystem v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/fxutil v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/hostname/validate v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/http v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/log v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/optional v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/pointer v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/startstop v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/statstracker v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/system v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/system/socket v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/winutil v0.60.1 // indirect
	github.com/DataDog/datadog-agent/pkg/version v0.60.1 // indirect
	github.com/DataDog/datadog-api-client-go/v2 v2.34.0 // indirect
	github.com/DataDog/datadog-go/v5 v5.6.0 // indirect
	github.com/DataDog/dd-sensitive-data-scanner/sds-go/go v0.0.0-20240816154533-f7f9beb53a42 // indirect
	github.com/DataDog/go-sqllexer v0.0.16 // indirect
	github.com/DataDog/go-tuf v1.1.0-0.5.2 // indirect
	github.com/DataDog/gohai v0.0.0-20230524154621-4316413895ee // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/inframetadata v0.22.0 // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes v0.22.0 // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/logs v0.22.0 // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/metrics v0.22.0 // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/quantile v0.22.0 // indirect
	github.com/DataDog/sketches-go v1.4.6 // indirect
	github.com/DataDog/viper v1.14.0 // indirect
	github.com/DataDog/zstd v1.5.6 // indirect
	github.com/GehirnInc/crypt v0.0.0-20200316065508-bb7000b8a962 // indirect
	github.com/GoogleCloudPlatform/grpc-gcp-go/grpcgcp v1.5.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.25.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/collector v0.49.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/collector/googlemanagedprometheus v0.49.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.25.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/extension/googleclientauthextension v0.49.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.49.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/IBM/sarama v1.44.0 // indirect
	github.com/JohnCGriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/Khan/genqlient v0.7.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ReneKroon/ttlcache/v2 v2.11.0 // indirect
	github.com/SAP/go-hdb v1.12.7 // indirect
	github.com/SermoDigital/jose v0.9.2-0.20180104203859-803625baeddc // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/aerospike/aerospike-client-go/v7 v7.8.0 // indirect
	github.com/alecthomas/participle/v2 v2.1.1 // indirect
	github.com/alecthomas/units v0.0.0-20240626203959-61d1e3462e30 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/aliyun/aliyun-log-go-sdk v0.1.83 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/antchfx/xmlquery v1.4.3 // indirect
	github.com/antchfx/xpath v1.3.3 // indirect
	github.com/apache/arrow/go/v15 v15.0.0 // indirect
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
	github.com/apache/arrow/go/v17 v17.0.0 // indirect
	github.com/apache/pulsar-client-go v0.14.0 // indirect
	github.com/apache/thrift v0.21.0 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/aws/aws-sdk-go-v2 v1.32.7 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.28.7 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.48 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.22 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.44 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.32.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.71.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.34.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.34.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.3 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/beevik/ntp v1.4.3 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.4.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.7.1 // indirect
	github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f // indirect
	github.com/briandowns/spinner v1.23.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/cilium/ebpf v0.11.0 // indirect
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e // indirect
	github.com/cloudfoundry/go-cfclient/v3 v3.0.0-alpha.9 // indirect
	github.com/cncf/xds/go v0.0.0-20240905190251-b4127c9b8d78 // indirect
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/containerd/cgroups/v3 v3.0.3 // indirect
	github.com/containerd/containerd/api v1.7.19 // indirect
	github.com/containerd/errdefs v0.1.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.14.3 // indirect
	github.com/containerd/ttrpc v1.2.5 // indirect
	github.com/coreos/go-oidc/v3 v3.11.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.5 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/digitalocean/godo v1.118.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v24.0.0+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v27.4.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ebitengine/purego v0.8.1 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.6.0 // indirect
	github.com/elastic/go-docappender/v2 v2.3.3 // indirect
	github.com/elastic/go-elasticsearch/v7 v7.17.10 // indirect
	github.com/elastic/go-elasticsearch/v8 v8.17.0 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/go-structform v0.0.12 // indirect
	github.com/elastic/go-sysinfo v1.7.1 // indirect
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/elastic/lunes v0.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.0 // indirect
	github.com/envoyproxy/go-control-plane v0.13.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.1.0 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.0 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/expr-lang/expr v1.16.9 // indirect
	github.com/facebook/time v0.0.0-20240510113249-fa89cc575891 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/getsentry/sentry-go v0.31.1 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-jose/go-jose/v4 v4.0.2 // indirect
	github.com/go-kit/kit v0.12.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-martini/martini v0.0.0-20170121215854-22fa46961aab // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-resty/resty/v2 v2.13.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/go-zookeeper/zk v1.0.3 // indirect
	github.com/gobuffalo/flect v1.0.2 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/gocql/gocql v1.7.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/cadvisor v0.51.0 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/go-github/v67 v67.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/licenseclassifier/v2 v2.0.0 // indirect
	github.com/google/pprof v0.0.0-20241023014458-598669927662 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/safetext v0.0.0-20220905092116-b49f7bc46da2 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gophercloud/gophercloud v1.13.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/gosnmp/gosnmp v1.38.0 // indirect
	github.com/grafana/loki/pkg/push v0.0.0-20240514112848-a1b1eeb09583 // indirect
	github.com/grafana/regexp v0.0.0-20240518133315-a468a5bfb3bc // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grobie/gomemcache v0.0.0-20230213081705-239240bbc445 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hamba/avro/v2 v2.25.1 // indirect
	github.com/hashicorp/consul/api v1.31.0 // indirect
	github.com/hashicorp/cronexpr v1.1.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20240717122358-3d93bd3778f3 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95 // indirect
	github.com/hetznercloud/hcloud-go/v2 v2.10.2 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/influxdata/influxdb-observability/common v0.5.12 // indirect
	github.com/influxdata/influxdb-observability/influx2otel v0.5.12 // indirect
	github.com/influxdata/influxdb-observability/otel2influx v0.5.12 // indirect
	github.com/influxdata/line-protocol/v2 v2.2.1 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.1.11 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jaegertracing/jaeger v1.64.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/leodido/go-syslog/v4 v4.2.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/lestrrat-go/strftime v1.1.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/linkedin/goavro/v2 v2.13.0 // indirect
	github.com/linode/linodego v1.37.0 // indirect
	github.com/logicmonitor/lm-data-sdk-go v1.3.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20240408141607-282e7b5d6b74 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/martini-contrib/render v0.0.0-20150707142108-ec18f8345a11 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-ieproxy v0.0.11 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	github.com/microsoft/ApplicationInsights-Go v0.4.4 // indirect
	github.com/microsoft/go-mssqldb v1.8.0 // indirect
	github.com/miekg/dns v1.1.61 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/spdystream v0.4.0 // indirect
	github.com/moby/sys/mountinfo v0.7.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/mongodb-forks/digest v1.1.0 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nginxinc/nginx-prometheus-exporter v0.11.0 // indirect
	github.com/oklog/ulid/v2 v2.1.0 // indirect
	github.com/onsi/ginkgo/v2 v2.20.2 // indirect
	github.com/onsi/gomega v1.34.2 // indirect
	github.com/open-telemetry/opamp-go v0.17.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampcustommessages v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/awsutil v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/containerinsight v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/cwlogs v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/k8s v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/metrics v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/proxy v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/xray v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/collectd v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/exp/metrics v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/rabbitmq v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sqlquery v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/datadog v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/topic v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/status v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azure v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azurelogs v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/loki v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/signalfx v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/skywalking v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.117.0 // indirect
	github.com/open-telemetry/otel-arrow v0.31.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.1.14 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/opensearch-project/opensearch-go/v2 v2.3.0 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/oschwald/geoip2-golang v1.11.0 // indirect
	github.com/oschwald/maxminddb-golang v1.13.0 // indirect
	github.com/otiai10/copy v1.14.0 // indirect
	github.com/outcaste-io/ristretto v0.2.1 // indirect
	github.com/ovh/go-ovh v1.6.0 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/philhofer/fwd v1.1.3-0.20240916144458-20a13a1f6b7c // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.7 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20220216144756-c35f1ee13d7c // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus-community/windows_exporter v0.27.2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/prometheus/prometheus v0.54.1 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/redis/go-redis/v9 v9.7.0 // indirect
	github.com/relvacode/iso8601 v1.6.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/samber/lo v1.38.1 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.29 // indirect
	github.com/scalyr/dataset-go v0.20.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.7.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shirou/gopsutil/v4 v4.24.12 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/sapm-proto v0.17.0 // indirect
	github.com/sijms/go-ora/v2 v2.8.22 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/snowflakedb/gosnowflake v1.12.0 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/solarwindscloud/apm-proto v1.0.8 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/stormcat24/protodep v0.1.8 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tedsuo/rata v1.0.0 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.1069 // indirect
	github.com/tg123/go-htpasswd v1.2.3 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/tinylru v1.2.1 // indirect
	github.com/tidwall/wal v1.1.8 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/tinylib/msgp v1.2.5 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/ua-parser/uap-go v0.0.0-20240611065828-3a4781585db6 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/vbatts/tar-split v0.11.3 // indirect
	github.com/vektah/gqlparser/v2 v2.5.20 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/vmware/go-vmware-nsxt v0.0.0-20230223012718-d31b8a1ca05e // indirect
	github.com/vmware/govmomi v0.47.0 // indirect
	github.com/vultr/govultr/v2 v2.17.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.elastic.co/apm/module/apmzap/v2 v2.6.2 // indirect
	go.elastic.co/apm/v2 v2.6.2 // indirect
	go.elastic.co/fastjson v1.4.0 // indirect
	go.etcd.io/bbolt v1.3.11 // indirect
	go.mongodb.org/atlas v0.37.0 // indirect
	go.mongodb.org/mongo-driver v1.17.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector v0.117.0 // indirect
	go.opentelemetry.io/collector/client v1.23.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.117.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.117.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.117.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.23.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.117.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.117.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.23.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.23.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.23.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.117.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.23.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.117.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.117.0 // indirect
	go.opentelemetry.io/collector/consumer v1.23.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.117.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.117.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.117.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.117.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.117.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.117.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.117.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.117.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.117.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.117.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.117.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.23.0 // indirect
	go.opentelemetry.io/collector/filter v0.117.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.117.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.117.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.117.0 // indirect
	go.opentelemetry.io/collector/pdata v1.23.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.117.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.117.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.117.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.117.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.117.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.117.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.117.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.117.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.117.0 // indirect
	go.opentelemetry.io/collector/scraper v0.117.0 // indirect
	go.opentelemetry.io/collector/scraper/scraperhelper v0.117.0 // indirect
	go.opentelemetry.io/collector/semconv v0.117.0 // indirect
	go.opentelemetry.io/collector/service v0.117.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.6.0 // indirect
	go.opentelemetry.io/contrib/config v0.10.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.31.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.57.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.31.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.56.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.7.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.54.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.7.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.32.0 // indirect
	go.opentelemetry.io/otel/log v0.8.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk v1.32.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.7.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.starlark.net v0.0.0-20230525235612-a134d8f9ddca // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/fx v1.22.2 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/genproto v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.2 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/evanphx/json-patch.v5 v5.6.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/zorkian/go-datadog-api.v2 v2.30.0 // indirect
	howett.net/plist v1.0.0 // indirect
	k8s.io/apiserver v0.31.3 // indirect
	k8s.io/component-base v0.31.3 // indirect
	k8s.io/gengo/v2 v2.0.0-20240228010128-51d4e06bde70 // indirect
	k8s.io/kubelet v0.31.3 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	skywalking.apache.org/repo/goapi v0.0.0-20240104145220-ba7202308dd4 // indirect
)
