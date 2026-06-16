package admin

import (
	"github.com/gin-gonic/gin"
	"rustdesk-api/global"
	"rustdesk-api/http/request/admin"
	"rustdesk-api/http/response"
	"rustdesk-api/service"
	"strconv"
)

type Group struct {
}

// Detail group
// @Tags group
// @Summary Group details
// @Description Group details
// @Accept  json
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} response.Response{data=model.Group}
// @Failure 500 {object} response.Response
// @Router /admin/group/detail/{id} [get]
// @Security token
func (ct *Group) Detail(c *gin.Context) {
	id := c.Param("id")
	iid, _ := strconv.Atoi(id)
	u := service.AllService.GroupService.InfoById(uint(iid))
	if u.Id > 0 {
		response.Success(c, u)
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
	return
}

// Create Create a group
// @Tags group
// @Summary Create a group
// @Description Create group
// @Accept  json
// @Produce  json
// @Param body body admin.GroupForm true "Group information"
// @Success 200 {object} response.Response{data=model.Group}
// @Failure 500 {object} response.Response
// @Router /admin/group/create [post]
// @Security token
func (ct *Group) Create(c *gin.Context) {
	f := &admin.GroupForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := f.ToGroup()
	err := service.AllService.GroupService.Create(u)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// List list
// @Tags group
// @Summary Group List
// @Description Group list
// @Accept  json
// @Produce  json
// @Param page query int false "page number"
// @Param page_size query int false "page size"
// @Success 200 {object} response.Response{data=model.GroupList}
// @Failure 500 {object} response.Response
// @Router /admin/group/list [get]
// @Security token
func (ct *Group) List(c *gin.Context) {
	query := &admin.PageQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.GroupService.List(query.Page, query.PageSize, nil)
	response.Success(c, res)
}

// Update Edit
// @Tags group
// @Summary Group Editor
// @Description Group editing
// @Accept  json
// @Produce  json
// @Param body body admin.GroupForm true "Group information"
// @Success 200 {object} response.Response{data=model.Group}
// @Failure 500 {object} response.Response
// @Router /admin/group/update [post]
// @Security token
func (ct *Group) Update(c *gin.Context) {
	f := &admin.GroupForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if f.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := f.ToGroup()
	err := service.AllService.GroupService.Update(u)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete Delete
// @Tags group
// @Summary Group deletion
// @Description Group delete
// @Accept  json
// @Produce  json
// @Param body body admin.GroupForm true "Group information"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/group/delete [post]
// @Security token
func (ct *Group) Delete(c *gin.Context) {
	f := &admin.GroupForm{}
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
	u := service.AllService.GroupService.InfoById(f.Id)
	if u.Id > 0 {
		err := service.AllService.GroupService.Delete(u)
		if err == nil {
			response.Success(c, nil)
			return
		}
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}
