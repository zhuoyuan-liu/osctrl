package handlers

import (
	"fmt"
	"net/http"

	"github.com/jmpsec/osctrl/pkg/config"
	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/jmpsec/osctrl/pkg/users"
	"github.com/jmpsec/osctrl/pkg/utils"
	"github.com/rs/zerolog/log"
)

// UserHandler - GET Handler for single JSON nodes
func (h *HandlersApi) UserHandler(w http.ResponseWriter, r *http.Request) {
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(config.ServiceAPI, settings.NoEnvironmentID), false)
	// Extract username
	usernameVar := r.PathValue("username")
	if usernameVar == "" {
		apiErrorResponse(w, "error with username", http.StatusBadRequest, nil)
		return
	}
	// Get user
	user, err := h.Users.Get(usernameVar)
	if err != nil {
		apiErrorResponse(w, "error getting user", http.StatusInternalServerError, nil)
		return
	}
	// Get context data and check access
	ctx := r.Context().Value(ContextKey(contextAPI)).(ContextValue)
	if !h.Users.CheckPermissions(ctx[ctxUser], users.AdminLevel, users.NoEnvironment) {
		apiErrorResponse(w, "no access", http.StatusForbidden, fmt.Errorf("attempt to use API by user %s", ctx[ctxUser]))
		return
	}
	// Serialize and serve JSON
	if h.Settings.DebugService(config.ServiceAPI) {
		log.Debug().Msgf("DebugService: Returned user %s", usernameVar)
	}
	utils.HTTPResponse(w, utils.JSONApplicationUTF8, http.StatusOK, user)
}

// UsersHandler - GET Handler for multiple JSON nodes
func (h *HandlersApi) UsersHandler(w http.ResponseWriter, r *http.Request) {
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(config.ServiceAPI, settings.NoEnvironmentID), false)
	// Get context data and check access
	ctx := r.Context().Value(ContextKey(contextAPI)).(ContextValue)
	if !h.Users.CheckPermissions(ctx[ctxUser], users.AdminLevel, users.NoEnvironment) {
		apiErrorResponse(w, "no access", http.StatusForbidden, fmt.Errorf("attempt to use API by user %s", ctx[ctxUser]))
		return
	}
	// Get users
	users, err := h.Users.All()
	if err != nil {
		apiErrorResponse(w, "error getting users", http.StatusInternalServerError, err)
		return
	}
	if len(users) == 0 {
		apiErrorResponse(w, "no users", http.StatusNotFound, nil)
		return
	}
	// Serialize and serve JSON
	if h.Settings.DebugService(config.ServiceAPI) {
		log.Debug().Msg("DebugService: Returned users")
	}
	utils.HTTPResponse(w, utils.JSONApplicationUTF8, http.StatusOK, users)
}
