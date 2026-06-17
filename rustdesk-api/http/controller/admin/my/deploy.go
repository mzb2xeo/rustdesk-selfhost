package my

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"rustdesk-api/global"
	"rustdesk-api/http/response"
	"rustdesk-api/model"
	"rustdesk-api/service"
)

type Deploy struct{}

type createDeployTokenForm struct {
	PasswordMode   string `json:"password_mode"`
	CustomPassword string `json:"custom_password"`
}

type revokeDeployTokenForm struct {
	Id uint `json:"id"`
}

// CreateToken issues a short-lived deploy token for automated client setup.
func (ct *Deploy) CreateToken(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	if u == nil || u.Id == 0 {
		response.Fail(c, 403, response.TranslateMsg(c, "NeedLogin"))
		return
	}

	form := &createDeployTokenForm{}
	_ = c.ShouldBindJSON(form)

	passwordMode := strings.TrimSpace(form.PasswordMode)
	if passwordMode == "" {
		passwordMode = model.DeployPasswordModeStructured
	}
	if passwordMode != model.DeployPasswordModeStructured && passwordMode != model.DeployPasswordModeCustom {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+"invalid password_mode")
		return
	}

	customPassword := strings.TrimSpace(form.CustomPassword)
	if passwordMode == model.DeployPasswordModeCustom {
		if utf8.RuneCountInString(customPassword) < 4 || utf8.RuneCountInString(customPassword) > 32 {
			response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+"custom_password length must be 4-32")
			return
		}
	} else {
		customPassword = ""
	}

	dt, err := service.AllService.DeployTokenService.Create(u.Id, passwordMode, customPassword)
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
		"password_mode":        dt.PasswordMode,
		"script_url":           scriptURL,
		"powershell_command":   powershellCommand,
		"download_run_command": downloadRunCommand,
	})
}

// ListTokens returns deploy tokens for the current user (masked, for monitoring).
func (ct *Deploy) ListTokens(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	if u == nil || u.Id == 0 {
		response.Fail(c, 403, response.TranslateMsg(c, "NeedLogin"))
		return
	}

	page := uint(1)
	pageSize := uint(20)
	if p := c.Query("page"); p != "" {
		if v, err := strconv.ParseUint(p, 10, 64); err == nil && v > 0 {
			page = uint(v)
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.ParseUint(ps, 10, 64); err == nil && v > 0 && v <= 100 {
			pageSize = uint(v)
		}
	}

	res := service.AllService.DeployTokenService.List(page, pageSize, u.Id)
	response.Success(c, res)
}

// RevokeToken force-revokes an active deploy token.
func (ct *Deploy) RevokeToken(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	if u == nil || u.Id == 0 {
		response.Fail(c, 403, response.TranslateMsg(c, "NeedLogin"))
		return
	}

	form := &revokeDeployTokenForm{}
	if err := c.ShouldBindJSON(form); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if form.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}

	if err := service.AllService.DeployTokenService.RevokeById(form.Id, u.Id); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
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
