package my

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"rustdesk-api/http/request/admin"
	"rustdesk-api/http/response"
	"rustdesk-api/service"
	"time"
)

type Peer struct {
}

// List list
// @Tags my device
// @Summary Device List
// @Description Device list
// @Accept  json
// @Produce  json
// @Param page query int false "page number"
// @Param page_size query int false "page size"
// @Param time_ago query int false "time"
// @Param id query string false "ID"
// @Param hostname query string false "hostname"
// @Param uuids query string false "uuids separated by commas"
// @Success 200 {object} response.Response{data=model.PeerList}
// @Failure 500 {object} response.Response
// @Router /admin/my/peer/list [get]
// @Security token
func (ct *Peer) List(c *gin.Context) {
	query := &admin.PeerQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	u := service.AllService.UserService.CurUser(c)
	res := service.AllService.PeerService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		tx.Where("user_id = ?", u.Id)
		if query.TimeAgo > 0 {
			lt := time.Now().Unix() - int64(query.TimeAgo)
			tx.Where("last_online_time < ?", lt)
		}
		if query.TimeAgo < 0 {
			lt := time.Now().Unix() + int64(query.TimeAgo)
			tx.Where("last_online_time > ?", lt)
		}
		if query.Id != "" {
			tx.Where("id like ?", "%"+query.Id+"%")
		}
		if query.Hostname != "" {
			tx.Where("hostname like ?", "%"+query.Hostname+"%")
		}
		if query.Uuids != "" {
			tx.Where("uuid in (?)", query.Uuids)
		}
	})
	response.Success(c, res)
}
