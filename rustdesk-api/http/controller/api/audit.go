package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"rustdesk-api/global"
	request "rustdesk-api/http/request/api"
	"rustdesk-api/http/response"
	"rustdesk-api/model"
	"rustdesk-api/service"
	"strconv"
	"time"
)

type Audit struct {
}

// AuditConn
// @Tags Audit
// @Summary Audit connection
// @Description Audit connection
// @Accept  json
// @Produce  json
// @Param body body request.AuditConnForm true "Audit connection"
// @Success 200 {string} string ""
// @Failure 500 {object} response.Response
// @Router /audit/conn [post]
func (a *Audit) AuditConn(c *gin.Context) {
	af := &request.AuditConnForm{}
	err := c.ShouldBindBodyWith(af, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	/*ttt := &gin.H{}
	c.ShouldBindBodyWith(ttt, binding.JSON)
	fmt.Println(ttt)*/
	ac := af.ToAuditConn()
	if af.Action == model.AuditActionNew {
		ac.Guid = uuid.New().String()
		service.AllService.AuditService.CreateAuditConn(ac)
		response.Success(c, ac.Guid)
		return
	} else if af.Action == model.AuditActionClose {
		ex := service.AllService.AuditService.InfoByPeerIdAndConnId(af.Id, af.ConnId)
		if ex.Id != 0 {
			ex.CloseTime = time.Now().Unix()
			service.AllService.AuditService.UpdateAuditConn(ex)
		}
	} else if af.Action == "" {
		ex := service.AllService.AuditService.InfoByPeerIdAndConnId(af.Id, af.ConnId)
		if ex.Id != 0 {
			up := &model.AuditConn{
				IdModel:   model.IdModel{Id: ex.Id},
				FromPeer:  ac.FromPeer,
				FromName:  ac.FromName,
				SessionId: ac.SessionId,
				Type:      ac.Type,
			}
			service.AllService.AuditService.UpdateAuditConn(up)
		}
	}
	response.Success(c, "")
}

// AuditFile
// @Tags Audit
// @Summary audit file
// @Description audit documents
// @Accept  json
// @Produce  json
// @Param body body request.AuditFileForm true "Audit file"
// @Success 200 {string} string ""
// @Failure 500 {object} response.Response
// @Router /audit/file [post]
func (a *Audit) AuditFile(c *gin.Context) {
	aff := &request.AuditFileForm{}
	err := c.ShouldBindBodyWith(aff, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	//ttt := &gin.H{}
	//c.ShouldBindBodyWith(ttt, binding.JSON)
	//fmt.Println(ttt)
	af := aff.ToAuditFile()
	service.AllService.AuditService.CreateAuditFile(af)
	response.Success(c, "")
}

func (a *Audit) AuditConnActive(c *gin.Context) {
	peerId := c.Query("id")
	sessionId := c.Query("session_id")
	connTypeStr := c.Query("conn_type")

	if peerId == "" || sessionId == "" {
		c.JSON(400, "Thiếu tham số id và session_id")
		return
	}

	connType, _ := strconv.Atoi(connTypeStr)

	currentUser := service.AllService.UserService.CurUser(c)
	if currentUser == nil || currentUser.Id == 0 {
		c.JSON(401, "Chưa xác thực")
		return
	}

	ac := &model.AuditConn{}
	err := global.DB.Where("peer_id = ? AND session_id = ? AND type = ? AND close_time = 0", peerId, sessionId, connType).Order("id DESC").First(ac).Error
	if err != nil {
		c.JSON(404, "Không tìm thấy kết nối")
		return
	}

	var count int64
	global.DB.Model(&model.Peer{}).Where("(id = ? OR id = ?) AND user_id = ?", ac.FromPeer, ac.PeerId, currentUser.Id).Count(&count)
	if count == 0 {
		c.JSON(403, "Không có quyền truy cập")
		return
	}

	c.JSON(200, ac.Guid)
}

func (a *Audit) UpdateAuditNote(c *gin.Context) {
	var form struct {
		Guid string `json:"guid" binding:"required"`
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}

	if len(form.Note) > 500 {
		response.Error(c, "Ghi chú quá dài (tối đa 500 ký tự)")
		return
	}

	currentUser := service.AllService.UserService.CurUser(c)
	if currentUser == nil || currentUser.Id == 0 {
		c.JSON(401, "Chưa xác thực")
		return
	}

	ac := &model.AuditConn{}
	err := global.DB.Where("guid = ?", form.Guid).First(ac).Error
	if err != nil {
		c.JSON(404, "Không tìm thấy thông tin kiểm tra kết nối")
		return
	}

	var count int64
	global.DB.Model(&model.Peer{}).Where("(id = ? OR id = ?) AND user_id = ?", ac.FromPeer, ac.PeerId, currentUser.Id).Count(&count)
	if count == 0 {
		c.JSON(403, "Không có quyền truy cập")
		return
	}

	ac.Note = form.Note
	err = service.AllService.AuditService.UpdateAuditConn(ac)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	response.Success(c, nil)
}

func (a *Audit) AuditAlarm(c *gin.Context) {
	af := &request.AuditAlarmForm{}
	err := c.ShouldBindJSON(af)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}

	var count int64
	global.DB.Model(&model.AuditAlarm{}).
		Where("peer_id = ? AND created_at > ?", af.Id, time.Now().Add(-10*time.Second)).
		Count(&count)
	if count > 0 {
		c.JSON(429, "Yêu cầu quá nhanh. Đã vượt quá giới hạn tần suất.")
		return
	}

	alarm := af.ToAuditAlarm(c.ClientIP())
	err = service.AllService.AuditService.CreateAuditAlarm(alarm)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "")
}
