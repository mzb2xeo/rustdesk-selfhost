package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"rustdesk-api/global"
	requstform "rustdesk-api/http/request/api"
	"rustdesk-api/http/response"
	"rustdesk-api/model"
	"rustdesk-api/service"
	"rustdesk-api/utils"
	"strings"
	"time"
)

type Index struct {
}

// Index Home Page
// @Tags Home Page
// @Summary Home Page
// @Description front page
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router / [get]
func (i *Index) Index(c *gin.Context) {
	response.Success(
		c,
		"Hello Gwen",
	)
}

// Heartbeat
// @Tags Home Page
// @Summary heartbeat
// @Description heartbeat
// @Accept  json
// @Produce  json
// @Success 200 {object} nil
// @Failure 500 {object} response.Response
// @Router /heartbeat [post]
func (i *Index) Heartbeat(c *gin.Context) {
	info := &requstform.PeerInfoInHeartbeat{}
	err := c.ShouldBindJSON(info)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if info.Uuid == "" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	peer := service.AllService.PeerService.FindById(info.Id)
	if peer == nil || peer.RowId == 0 {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	//If it is within 40s, it will not be updated.
	if time.Now().Unix()-peer.LastOnlineTime >= 30 {
		upp := &model.Peer{RowId: peer.RowId, LastOnlineTime: time.Now().Unix(), LastOnlineIp: c.ClientIP()}
		service.AllService.PeerService.Update(upp)
	}
	c.JSON(http.StatusOK, gin.H{})
}

// Version version
// @Tags Home Page
// @Summary version
// @Description Version
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /version [get]
func (i *Index) Version(c *gin.Context) {
	//Read resources/version file
	v := service.AllService.AppService.GetAppVersion()
	response.Success(
		c,
		v,
	)
}

// DeployPowershell returns automated powershell configuration script (deploy token required).
func (i *Index) DeployPowershell(c *gin.Context) {
	deployToken := strings.TrimSpace(c.Query("deploy_token"))
	if deployToken == "" {
		c.String(http.StatusBadRequest, "deploy_token is required")
		return
	}
	dt, err := service.AllService.DeployTokenService.FindValid(deployToken)
	if err != nil {
		c.String(http.StatusUnauthorized, "invalid or expired deploy token")
		return
	}

	apiServer := resolvePublicApiServer(c)
	idServer := resolvePublicIdServer(c)
	relayServer := resolvePublicRelayServer(c)
	global.Config.Rustdesk.LoadKeyFile()
	key := global.Config.Rustdesk.Key
	configString := utils.EncodeRustDeskConfig(idServer, relayServer, apiServer, key)

	passwordMode := dt.PasswordMode
	if passwordMode == "" {
		passwordMode = model.DeployPasswordModeStructured
	}

	script := loadPowershellTemplate()
	script = strings.ReplaceAll(script, "{{.DeployToken}}", deployToken)
	script = strings.ReplaceAll(script, "{{.ApiUrl}}", apiServer)
	script = strings.ReplaceAll(script, "{{.IdServer}}", idServer)
	script = strings.ReplaceAll(script, "{{.RelayServer}}", relayServer)
	script = strings.ReplaceAll(script, "{{.Key}}", key)
	script = strings.ReplaceAll(script, "{{.ConfigString}}", configString)
	script = strings.ReplaceAll(script, "{{.PasswordMode}}", passwordMode)
	script = strings.ReplaceAll(script, "{{.CustomPassword}}", escapePowerShellSingleQuoted(dt.CustomPassword))

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, script)
}

func loadPowershellTemplate() string {
	paths := []string{
		filepath.Join("data", "templates", "deploy-host.ps1"),
		filepath.Join(global.Config.Gin.ResourcesPath, "templates", "deploy-host.ps1"),
		filepath.Join("resources", "templates", "deploy-host.ps1"),
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil && len(content) > 0 {
			return string(content)
		}
	}
	global.Logger.Error("deploy-host.ps1 template not found")
	return "Write-Error 'Deploy template missing on server'; exit 1`n"
}

// DeployRevoke consumes a deploy token after successful setup.
func (i *Index) DeployRevoke(c *gin.Context) {
	authType, _ := c.Get("authType")
	if authType != "deploy" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	token, _ := c.Get("token")
	if token == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	_ = service.AllService.DeployTokenService.Consume(token.(string))
	c.JSON(http.StatusOK, gin.H{})
}

func resolvePublicApiServer(c *gin.Context) string {
	apiServer := global.Config.Rustdesk.ApiServer
	if apiServer == "" || strings.Contains(apiServer, "127.0.0.1") || strings.Contains(apiServer, "localhost") {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		apiServer = scheme + "://" + c.Request.Host
	}
	return strings.TrimRight(apiServer, "/")
}

func resolvePublicIdServer(c *gin.Context) string {
	idServer := global.Config.Rustdesk.IdServer
	if idServer == "" {
		host := c.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		idServer = host + ":21116"
	}
	return idServer
}

func resolvePublicRelayServer(c *gin.Context) string {
	relayServer := global.Config.Rustdesk.RelayServer
	if relayServer == "" {
		host := c.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		relayServer = host + ":21117"
	}
	return relayServer
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
