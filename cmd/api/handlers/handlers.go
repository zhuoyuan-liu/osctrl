package handlers

import (
	"github.com/jmpsec/osctrl/pkg/cache"
	"github.com/jmpsec/osctrl/pkg/carves"
	"github.com/jmpsec/osctrl/pkg/config"
	"github.com/jmpsec/osctrl/pkg/environments"
	"github.com/jmpsec/osctrl/pkg/logging"
	"github.com/jmpsec/osctrl/pkg/nodes"
	"github.com/jmpsec/osctrl/pkg/queries"
	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/jmpsec/osctrl/pkg/tags"
	"github.com/jmpsec/osctrl/pkg/users"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const errorContent = "❌"
const okContent = "✅"

type HandlersApi struct {
	DB              *gorm.DB
	Users           *users.UserManager
	Tags            *tags.TagManager
	Envs            *environments.EnvManager
	Nodes           *nodes.NodeManager
	Queries         *queries.Queries
	Carves          *carves.Carves
	Settings        *settings.Settings
	RedisCache      *cache.RedisManager
	ServiceVersion  string
	ServiceName     string
	ApiConfig       *config.JSONConfigurationService
	DebugHTTP       *zerolog.Logger
	DebugHTTPConfig *config.DebugHTTPConfiguration
}

type HandlersOption func(*HandlersApi)

func WithDB(db *gorm.DB) HandlersOption {
	return func(h *HandlersApi) {
		h.DB = db
	}
}

func WithUsers(users *users.UserManager) HandlersOption {
	return func(h *HandlersApi) {
		h.Users = users
	}
}

func WithTags(tags *tags.TagManager) HandlersOption {
	return func(h *HandlersApi) {
		h.Tags = tags
	}
}

func WithEnvs(envs *environments.EnvManager) HandlersOption {
	return func(h *HandlersApi) {
		h.Envs = envs
	}
}

func WithNodes(nodes *nodes.NodeManager) HandlersOption {
	return func(h *HandlersApi) {
		h.Nodes = nodes
	}
}

func WithQueries(queries *queries.Queries) HandlersOption {
	return func(h *HandlersApi) {
		h.Queries = queries
	}
}

func WithCarves(carves *carves.Carves) HandlersOption {
	return func(h *HandlersApi) {
		h.Carves = carves
	}
}

func WithSettings(settings *settings.Settings) HandlersOption {
	return func(h *HandlersApi) {
		h.Settings = settings
	}
}

func WithCache(rds *cache.RedisManager) HandlersOption {
	return func(h *HandlersApi) {
		h.RedisCache = rds
	}
}

func WithVersion(version string) HandlersOption {
	return func(h *HandlersApi) {
		h.ServiceVersion = version
	}
}

func WithName(name string) HandlersOption {
	return func(h *HandlersApi) {
		h.ServiceName = name
	}
}

func WithDebugHTTP(cfg *config.DebugHTTPConfiguration) HandlersOption {
	return func(h *HandlersApi) {
		h.DebugHTTPConfig = cfg
		h.DebugHTTP = nil
		if cfg.Enabled {
			l, err := logging.CreateDebugHTTP(cfg.File, logging.LumberjackConfig{
				MaxSize:    25,
				MaxBackups: 5,
				MaxAge:     10,
				Compress:   true,
			})
			if err != nil {
				log.Err(err).Msg("error creating debug HTTP logger")
				l = nil
				h.DebugHTTPConfig.Enabled = false
			}
			h.DebugHTTP = l
		}
	}
}

// CreateHandlersApi to initialize the Admin handlers struct
func CreateHandlersApi(opts ...HandlersOption) *HandlersApi {
	h := &HandlersApi{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
