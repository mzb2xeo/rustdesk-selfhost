package admin

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"rustdesk-api/global"
	"rustdesk-api/http/request/admin"
	"rustdesk-api/http/response"
	"rustdesk-api/service"
)

type ShareRecord struct {
}

// List list
// @Tags share records
// @Summary Share record list
// @Description Share record list
// @Accept  json
// @Produce  json
// @Param user_id query int false "user ID"
// @Param page query int false "page number"
// @Param page_size query int false "page size"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/share_record/list [get]
// @Security token
func (sr *ShareRecord) List(c *gin.Context) {
	query := &admin.ShareRecordQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.ShareRecordService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.UserId > 0 {
			tx.Where("user_id = ?", query.UserId)
		}
	})
	response.Success(c, res)
}

// Delete Delete
// @Tags share records
// @Summary Sharing record deleted
// @Description Delete sharing history
// @Accept  json
// @Produce  json
// @Param body body admin.ShareRecordForm true "Share record information"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/share_record/delete [post]
// @Security token
func (sr *ShareRecord) Delete(c *gin.Context) {
	f := &admin.ShareRecordForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	id := f.Id
	errList := global.Validator.ValidVar(c, id, "required,gt=0")
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	i := service.AllService.ShareRecordService.InfoById(f.Id)
	if i.Id > 0 {
		err := service.AllService.ShareRecordService.Delete(i)
		if err == nil {
			response.Success(c, nil)
			return
		}
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}

// BatchDelete Batch delete
// @Tags share records
// @Summary Batch sharing records
// @Description Share records in batches
// @Accept  json
// @Produce  json
// @Param body body admin.PeerShareRecordBatchDeleteForm true "id"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/share_record/batchDelete [post]
// @Security token
func (sr *ShareRecord) BatchDelete(c *gin.Context) {
	f := &admin.PeerShareRecordBatchDeleteForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if len(f.Ids) == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	err := service.AllService.ShareRecordService.BatchDelete(f.Ids)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}
