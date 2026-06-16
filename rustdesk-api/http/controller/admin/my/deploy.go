package my

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"rustdesk-api/global"
	"rustdesk-api/http/response"
	"rustdesk-api/model"
	"rustdesk-api/service"
)

type Deploy struct{}

// CreateToken issues a short-lived deploy token for automated client setup.
func (ct *Deploy) CreateToken(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	if u == nil || u.Id == 0 {
		response.Fail(c, 403, response.TranslateMsg(c, "NeedLogin"))
		return
	}

	dt, err := service.AllService.DeployTokenService.Create(u.Id)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}

	apiServer := resolveApiServer(c)
	scriptURL := fmt.Sprintf("%s/api/deploy/powershell?deploy_token=%s", apiServer, dt.Token)
	powershellCommand := fmt.Sprintf(
		`powershell -ExecutionPolicy Bypass -Command "[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; iex (Invoke-WebRequest -UseBasicParsing -Uri '%s').Content"`,
		scriptURL,
	)
	downloadRunCommand := fmt.Sprintf(
		"powershell -NoProfile -ExecutionPolicy Bypass -Command \"`$u='%s'; `$p=Join-Path `$env:TEMP 'rustdesk-deploy.ps1'; [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -UseBasicParsing -Uri `$u -OutFile `$p; powershell -NoProfile -ExecutionPolicy Bypass -File `$p\"",
		scriptURL,
	)

	response.Success(c, gin.H{
		"deploy_token":         dt.Token,
		"expires_at":           dt.ExpiresAt,
		"expires_in":           model.DeployTokenTTLSeconds,
		"script_url":           scriptURL,
		"powershell_command":   powershellCommand,
		"download_run_command": downloadRunCommand,
	})
}

func resolveApiServer(c *gin.Context) string {
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
