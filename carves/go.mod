module github.com/jmpsec/osctrl/carves

go 1.23

replace github.com/jmpsec/osctrl/nodes => ../nodes

replace github.com/jmpsec/osctrl/queries => ../queries

replace github.com/jmpsec/osctrl/types => ../types

replace github.com/jmpsec/osctrl/settings => ../settings

replace github.com/jmpsec/osctrl/utils => ../utils

require (
	github.com/jinzhu/gorm v1.9.16
	github.com/jmpsec/osctrl/nodes v0.3.9 // indirect
	github.com/jmpsec/osctrl/queries v0.3.9 // indirect
	github.com/jmpsec/osctrl/types v0.0.0-20240926152724-6f0cf5eace97
)

require (
	github.com/aws/aws-sdk-go-v2 v1.31.0
	github.com/aws/aws-sdk-go-v2/config v1.27.38
	github.com/aws/aws-sdk-go-v2/credentials v1.17.36
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.24
	github.com/aws/aws-sdk-go-v2/service/s3 v1.63.2
	github.com/jmpsec/osctrl/settings v0.0.0-20240926110606-74392bf45499
	github.com/jmpsec/osctrl/utils v0.0.0-20240926152724-6f0cf5eace97
	github.com/spf13/viper v1.19.0
	gorm.io/gorm v1.25.12
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.23.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.27.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.31.2 // indirect
	github.com/aws/smithy-go v1.21.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/rs/zerolog v1.33.0
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
