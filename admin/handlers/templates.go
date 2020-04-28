package handlers

import (
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/jmpsec/osctrl/admin/sessions"
	"github.com/jmpsec/osctrl/carves"
	"github.com/jmpsec/osctrl/environments"
	"github.com/jmpsec/osctrl/settings"
	"github.com/jmpsec/osctrl/users"
	"github.com/jmpsec/osctrl/utils"
)

const (
	templatesFilesFolder string = "tmpl_admin"
)

// TemplateFiles for building UI layout
type TemplateFiles struct {
	filepaths []string
}

// Helper to prepare template metadata
func (h *HandlersAdmin) TemplateMetadata(ctx sessions.ContextValue, version string) TemplateMetadata {
	return TemplateMetadata{
		Username:       ctx[sessions.CtxUser],
		Level:          ctx[sessions.CtxLevel],
		CSRFToken:      ctx[sessions.CtxCSRF],
		Service:        "osctrl-admin",
		Version:        version,
		TLSDebug:       h.Settings.DebugService(settings.ServiceTLS),
		AdminDebug:     h.Settings.DebugService(settings.ServiceAdmin),
		APIDebug:       h.Settings.DebugService(settings.ServiceAPI),
		AdminDebugHTTP: h.Settings.DebugHTTP(settings.ServiceAdmin),
		APIDebugHTTP:   h.Settings.DebugHTTP(settings.ServiceAPI),
	}
}

// NewTemplateFiles defines based on layout and default static pages
func NewTemplateFiles(base string, layoutFilename string) *TemplateFiles {
	paths := []string{
		base + "/" + layoutFilename,
		base + "/components/page-head.html",
		base + "/components/page-js.html",
		base + "/components/page-header.html",
		base + "/components/page-aside-left.html",
		base + "/components/page-aside-right.html",
		base + "/components/page-modals.html",
	}
	tf := TemplateFiles{filepaths: paths}
	return &tf
}

// LoginHandler for login page for GET requests
func (h *HandlersAdmin) LoginHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Prepare template
	t, err := template.ParseFiles(
		templatesFilesFolder+"/login.html",
		templatesFilesFolder+"/components/page-head.html",
		templatesFilesFolder+"/components/page-js.html")
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting login template: %v", err)
		return
	}
	// Prepare template data
	templateData := LoginTemplateData{
		Title:   "Login to osctrl",
		Project: "osctrl",
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Login template served")
	}
	h.Inc(metricAdminOK)
}

// EnvironmentHandler for environment view of the table
func (h *HandlersAdmin) EnvironmentHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	vars := mux.Vars(r)
	// Extract environment
	env, ok := vars["environment"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting environment")
		return
	}
	// Check if environment is valid
	if !h.Envs.Exists(env) {
		h.Inc(metricAdminErr)
		log.Printf("error unknown environment (%s)", env)
		return
	}
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.EnvLevel, env) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricTokenErr)
		return
	}
	// Extract target
	// FIXME verify target
	target, ok := vars["target"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting target")
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "table.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)

	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Prepare template data
	templateData := TableTemplateData{
		Title:        "Nodes in " + env,
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Selector:     "environment",
		SelectorName: env,
		Target:       target,
		Environments: envAll,
		Platforms:    platforms,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Environment table template served")
	}
	h.Inc(metricAdminOK)
}

// PlatformHandler for platform view of the table
func (h *HandlersAdmin) PlatformHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	vars := mux.Vars(r)
	// Extract platform
	// FIXME verify platform
	platform, ok := vars["platform"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting platform")
		return
	}
	// Extract target
	// FIXME verify target
	target, ok := vars["target"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting target")
		return
	}
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.AdminLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricTokenErr)
		return
	}
	// Prepare template
	t, err := template.ParseFiles(
		templatesFilesFolder+"/table.html",
		templatesFilesFolder+"/components/page-head.html",
		templatesFilesFolder+"/components/page-js.html",
		templatesFilesFolder+"/components/page-aside-right.html",
		templatesFilesFolder+"/components/page-aside-left.html",
		templatesFilesFolder+"/components/page-header.html",
		templatesFilesFolder+"/components/page-modals.html")
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Prepare template data
	templateData := TableTemplateData{
		Title:        "Nodes in " + platform,
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Selector:     "platform",
		SelectorName: platform,
		Target:       target,
		Environments: envAll,
		Platforms:    platforms,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Platform table template served")
	}
	h.Inc(metricAdminOK)
}

// QueryRunGETHandler for GET requests to run queries
func (h *HandlersAdmin) QueryRunGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.QueryLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Prepare template
	t, err := template.ParseFiles(
		templatesFilesFolder+"/queries-run.html",
		templatesFilesFolder+"/components/page-head.html",
		templatesFilesFolder+"/components/page-js.html",
		templatesFilesFolder+"/components/page-aside-right.html",
		templatesFilesFolder+"/components/page-aside-left.html",
		templatesFilesFolder+"/components/page-header.html",
		templatesFilesFolder+"/components/page-modals.html")
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get all nodes
	nodes, err := h.Nodes.Gets("active", h.Settings.InactiveHours())
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting all nodes: %v", err)
		return
	}
	// Convert to list of UUIDs and Hosts
	// FIXME if the number of nodes is big, this may cause issues loading the page
	var uuids, hosts []string
	for _, n := range nodes {
		uuids = append(uuids, n.UUID)
		hosts = append(hosts, n.Localname)
	}
	// Prepare template data
	templateData := QueryRunTemplateData{
		Title:         "Query osquery Nodes",
		Metadata:      h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments:  envAll,
		Platforms:     platforms,
		UUIDs:         uuids,
		Hosts:         hosts,
		Tables:        h.OsqueryTables,
		TablesVersion: osqueryTablesVersion,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Query run template served")
	}
	h.Inc(metricAdminOK)
}

// QueryListGETHandler for GET requests to queries
func (h *HandlersAdmin) QueryListGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.QueryLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "queries.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Prepare template data
	templateData := QueryTableTemplateData{
		Title:        "All on-demand queries",
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments: envAll,
		Platforms:    platforms,
		Target:       "all",
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Query list template served")
	}
	h.Inc(metricAdminOK)
}

// CarvesRunGETHandler for GET requests to run file carves
func (h *HandlersAdmin) CarvesRunGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.CarveLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "carves-run.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get all nodes
	nodes, err := h.Nodes.Gets("active", h.Settings.InactiveHours())
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting all nodes: %v", err)
		return
	}
	// Convert to list of UUIDs and Hosts
	// FIXME if the number of nodes is big, this may cause issues loading the page
	var uuids, hosts []string
	for _, n := range nodes {
		uuids = append(uuids, n.UUID)
		hosts = append(hosts, n.Localname)
	}
	// Prepare template data
	templateData := CarvesRunTemplateData{
		Title:         "Query osquery Nodes",
		Metadata:      h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments:  envAll,
		Platforms:     platforms,
		UUIDs:         uuids,
		Hosts:         hosts,
		Tables:        h.OsqueryTables,
		TablesVersion: osqueryTablesVersion,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Query run template served")
	}
	h.Inc(metricAdminOK)
}

// CarvesListGETHandler for GET requests to carves
func (h *HandlersAdmin) CarvesListGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.CarveLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "carves.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Prepare template data
	templateData := CarvesTableTemplateData{
		Title:        "All carved files",
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments: envAll,
		Platforms:    platforms,
		Target:       "all",
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Carve list template served")
	}
	h.Inc(metricAdminOK)
}

// QueryLogsHandler for GET requests to see query results by name
func (h *HandlersAdmin) QueryLogsHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.QueryLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	vars := mux.Vars(r)
	// Extract name
	name, ok := vars["name"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting name")
		return
	}
	// Custom functions to handle formatting
	funcMap := template.FuncMap{
		"queryResultLink":  h.queryResultLink,
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "queries-logs.html").filepaths
	t, err := template.New("queries-logs.html").Funcs(funcMap).ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get query by name
	query, err := h.Queries.Get(name)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting query %v", err)
		return
	}
	// Get query targets
	targets, err := h.Queries.GetTargets(name)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting targets %v", err)
		return
	}
	// Prepare template data
	templateData := QueryLogsTemplateData{
		Title:        "Query logs " + query.Name,
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments: envAll,
		Platforms:    platforms,
		Query:        query,
		QueryTargets: targets,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Query logs template served")
	}
	h.Inc(metricAdminOK)
}

// CarvesDetailsHandler for GET requests to see carves details by name
func (h *HandlersAdmin) CarvesDetailsHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.CarveLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	vars := mux.Vars(r)
	// Extract name
	name, ok := vars["name"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting name")
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "carves-details.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}

	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get query by name
	query, err := h.Queries.Get(name)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting query %v", err)
		return
	}
	// Get query targets
	targets, err := h.Queries.GetTargets(name)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting targets %v", err)
		return
	}
	// Get carves for this query
	queryCarves, err := h.Carves.GetByQuery(name)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting carve %v", err)
		return
	}
	// Get carve blocks by carve
	blocks := make(map[string][]carves.CarvedBlock)
	for _, c := range queryCarves {
		bs, err := h.Carves.GetBlocks(c.SessionID)
		if err != nil {
			h.Inc(metricAdminErr)
			log.Printf("error getting carve blocks %v", err)
			break
		}
		blocks[c.SessionID] = bs
	}
	// Prepare template data
	templateData := CarvesDetailsTemplateData{
		Title:        "Carve details " + query.Name,
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments: envAll,
		Platforms:    platforms,
		Query:        query,
		QueryTargets: targets,
		Carves:       queryCarves,
		CarveBlocks:  blocks,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Carve details template served")
	}
	h.Inc(metricAdminOK)
}

// ConfGETHandler for GET requests for /conf
func (h *HandlersAdmin) ConfGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	vars := mux.Vars(r)
	// Extract environment
	envVar, ok := vars["environment"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("environment is missing")
		return
	}
	// Check if environment is valid
	if !h.Envs.Exists(envVar) {
		h.Inc(metricAdminErr)
		log.Printf("error unknown environment (%s)", envVar)
		return
	}
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.EnvLevel, envVar) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "conf.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting conf template: %v", err)
		return
	}
	// Get stats for all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get stats for all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get configuration JSON
	env, err := h.Envs.Get(envVar)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environment %v", err)
		return
	}
	// Prepare template data
	templateData := ConfTemplateData{
		Title:        envVar + " Configuration",
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Environment:  env,
		Environments: envAll,
		Platforms:    platforms,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Conf template served")
	}
	h.Inc(metricAdminOK)
}

// EnrollGETHandler for GET requests for /enroll
func (h *HandlersAdmin) EnrollGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	vars := mux.Vars(r)
	// Extract environment
	envVar, ok := vars["environment"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("environment is missing")
		return
	}
	// Check if environment is valid
	if !h.Envs.Exists(envVar) {
		h.Inc(metricAdminErr)
		log.Printf("error unknown environment (%s)", envVar)
		return
	}
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.EnvLevel, envVar) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "enroll.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting enroll template: %v", err)
		return
	}
	// Get stats for all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get stats for all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get configuration JSON
	env, err := h.Envs.Get(envVar)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environment %v", err)
		return
	}
	// Prepare template data
	shellQuickAdd, _ := environments.QuickAddOneLinerShell(env)
	powershellQuickAdd, _ := environments.QuickAddOneLinerPowershell(env)
	shellQuickRemove, _ := environments.QuickRemoveOneLinerShell(env)
	powershellQuickRemove, _ := environments.QuickRemoveOneLinerPowershell(env)
	templateData := EnrollTemplateData{
		Title:                 envVar + " Enroll",
		Metadata:              h.TemplateMetadata(ctx, h.ServiceVersion),
		EnvName:               envVar,
		EnrollExpiry:          strings.ToUpper(utils.InFutureTime(env.EnrollExpire)),
		EnrollExpired:         environments.IsItExpired(env.EnrollExpire),
		RemoveExpiry:          strings.ToUpper(utils.InFutureTime(env.RemoveExpire)),
		RemoveExpired:         environments.IsItExpired(env.RemoveExpire),
		QuickAddShell:         shellQuickAdd,
		QuickRemoveShell:      shellQuickRemove,
		QuickAddPowershell:    powershellQuickAdd,
		QuickRemovePowershell: powershellQuickRemove,
		Secret:                env.Secret,
		Flags:                 env.Flags,
		Certificate:           env.Certificate,
		Environments:          envAll,
		Platforms:             platforms,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Enroll template served")
	}
	h.Inc(metricAdminOK)
}

// NodeHandler for node view
func (h *HandlersAdmin) NodeHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	vars := mux.Vars(r)
	// Extract uuid
	uuid, ok := vars["uuid"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting uuid")
		return
	}
	// Custom functions to handle formatting
	funcMap := template.FuncMap{
		"pastFutureTimes": utils.PastFutureTimes,
		"jsonRawIndent":   jsonRawIndent,
		"statusLogsLink":  h.statusLogsLink,
		"resultLogsLink":  h.resultLogsLink,
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "node.html").filepaths
	t, err := template.New("node.html").Funcs(funcMap).ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting table template: %v", err)
		return
	}
	// Get node by UUID
	node, err := h.Nodes.GetByUUID(uuid)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting node %v", err)
		return
	}
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.EnvLevel, node.Environment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Get all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments%v", err)
		return
	}
	// Get all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Prepare template data
	templateData := NodeTemplateData{
		Title:        "Node View " + node.Hostname,
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Node:         node,
		Environments: envAll,
		Platforms:    platforms,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Node template served")
	}
	h.Inc(metricAdminOK)
}

// EnvsGETHandler for GET requests for /env
func (h *HandlersAdmin) EnvsGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.AdminLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "environments.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments template: %v", err)
		return
	}
	// Get stats for all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get stats for all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Prepare template data
	templateData := EnvironmentsTemplateData{
		Title:        "Manage environments",
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments: envAll,
		Platforms:    platforms,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Environments template served")
	}
	h.Inc(metricAdminOK)
}

// SettingsGETHandler for GET requests for /settings
func (h *HandlersAdmin) SettingsGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	vars := mux.Vars(r)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.AdminLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Extract service
	serviceVar, ok := vars["service"]
	if !ok {
		h.Inc(metricAdminErr)
		log.Println("error getting service")
		return
	}
	// Verify service
	if !checkTargetService(serviceVar) {
		h.Inc(metricAdminErr)
		log.Printf("error unknown service (%s)", serviceVar)
		return
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "settings.html").filepaths
	t, err := template.ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments template: %v", err)
		return
	}
	// Get stats for all environments
	envAll, err := h.Envs.All()
	if err != nil {
		log.Printf("error getting environments %v", err)
		return
	}
	// Get stats for all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get setting values
	_settings, err := h.Settings.RetrieveValues(serviceVar)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting settings: %v", err)
		return
	}
	// Get JSON values
	svcJSON, err := h.Settings.RetrieveAllJSON(serviceVar)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting JSON values: %v", err)
	}
	// Prepare template data
	templateData := SettingsTemplateData{
		Title:           "Manage settings",
		Metadata:        h.TemplateMetadata(ctx, h.ServiceVersion),
		Service:         serviceVar,
		Environments:    envAll,
		Platforms:       platforms,
		CurrentSettings: _settings,
		ServiceConfig:   toJSONConfigurationService(svcJSON),
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Settings template served")
	}
	h.Inc(metricAdminOK)
}

// UsersGETHandler for GET requests for /users
func (h *HandlersAdmin) UsersGETHandler(w http.ResponseWriter, r *http.Request) {
	h.Inc(metricAdminReq)
	utils.DebugHTTPDump(r, h.Settings.DebugHTTP(settings.ServiceAdmin), false)
	// Get context data
	ctx := r.Context().Value(sessions.ContextKey("session")).(sessions.ContextValue)
	// Check permissions
	if !h.Users.CheckPermissions(ctx[sessions.CtxUser], users.AdminLevel, users.NoEnvironment) {
		log.Printf("%s has insuficient permissions", ctx[sessions.CtxUser])
		h.Inc(metricAdminErr)
		return
	}
	// Custom functions to handle formatting
	funcMap := template.FuncMap{
		"pastFutureTimes": utils.PastFutureTimes,
		"inFutureTime":    utils.InFutureTime,
	}
	// Prepare template
	tempateFiles := NewTemplateFiles(templatesFilesFolder, "users.html").filepaths
	t, err := template.New("users.html").Funcs(funcMap).ParseFiles(tempateFiles...)
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments template: %v", err)
		return
	}
	// Get stats for all environments
	envAll, err := h.Envs.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting environments %v", err)
		return
	}
	// Get stats for all platforms
	platforms, err := h.Nodes.GetAllPlatforms()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting platforms: %v", err)
		return
	}
	// Get current users
	users, err := h.Users.All()
	if err != nil {
		h.Inc(metricAdminErr)
		log.Printf("error getting users: %v", err)
		return
	}
	// Prepare template data
	templateData := UsersTemplateData{
		Title:        "Manage users",
		Metadata:     h.TemplateMetadata(ctx, h.ServiceVersion),
		Environments: envAll,
		Platforms:    platforms,
		CurrentUsers: users,
	}
	if err := t.Execute(w, templateData); err != nil {
		h.Inc(metricAdminErr)
		log.Printf("template error %v", err)
		return
	}
	if h.Settings.DebugService(settings.ServiceAdmin) {
		log.Println("DebugService: Users template served")
	}
	h.Inc(metricAdminOK)
}
