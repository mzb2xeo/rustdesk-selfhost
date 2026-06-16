package admin

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"rustdesk-api/global"
	"rustdesk-api/http/request/admin"
	"rustdesk-api/http/response"
	"rustdesk-api/model"
	"rustdesk-api/service"
	"strconv"
)

type AddressBookCollection struct {
}

// Detail address book name
// @Tags address book name
// @Summary Address book name details
// @Description Address book name details
// @Accept  json
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} response.Response{data=model.AddressBookCollection}
// @Failure 500 {object} response.Response
// @Router /admin/address_book_collection/detail/{id} [get]
// @Security token
func (abc *AddressBookCollection) Detail(c *gin.Context) {
	id := c.Param("id")
	iid, _ := strconv.Atoi(id)
	t := service.AllService.AddressBookService.CollectionInfoById(uint(iid))
	if t.Id > 0 {
		response.Success(c, t)
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
	return
}

// Create Create address book name
// @Tags address book name
// @Summary Create address book name
// @Description Create address book name
// @Accept  json
// @Produce  json
// @Param body body model.AddressBookCollection true "Address Book Name Information"
// @Success 200 {object} response.Response{data=model.AddressBookCollection}
// @Failure 500 {object} response.Response
// @Router /admin/address_book_collection/create [post]
// @Security token
func (abc *AddressBookCollection) Create(c *gin.Context) {
	f := &model.AddressBookCollection{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	if f.UserId == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	t := f
	err := service.AllService.AddressBookService.CreateCollection(t)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// List list
// @Tags address book name
// @Summary List of address book names
// @Description Address book name list
// @Accept  json
// @Produce  json
// @Param page query int false "page number"
// @Param page_size query int false "page size"
// @Param is_my query int false "Is it mine"
// @Param user_id query int false "userid"
// @Success 200 {object} response.Response{data=model.AddressBookCollectionList}
// @Failure 500 {object} response.Response
// @Router /admin/address_book_collection/list [get]
// @Security token
func (abc *AddressBookCollection) List(c *gin.Context) {
	query := &admin.AddressBookCollectionQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.AddressBookService.ListCollection(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.UserId > 0 {
			tx.Where("user_id = ?", query.UserId)
		}
	})
	response.Success(c, res)
}

// Update Edit
// @Tags address book name
// @Summary Address book name editing
// @Description Address book name editing
// @Accept  json
// @Produce  json
// @Param body body model.AddressBookCollection true "Address Book Name Information"
// @Success 200 {object} response.Response{data=model.AddressBookCollection}
// @Failure 500 {object} response.Response
// @Router /admin/address_book_collection/update [post]
// @Security token
func (abc *AddressBookCollection) Update(c *gin.Context) {
	f := &model.AddressBookCollection{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	if f.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	t := f //f.ToAddressBookCollection()
	err := service.AllService.AddressBookService.UpdateCollection(t)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete Delete
// @Tags address book name
// @Summary Address book name deletion
// @Description Address book name deletion
// @Accept  json
// @Produce  json
// @Param body body model.AddressBookCollection true "Address Book Name Information"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/address_book_collection/delete [post]
// @Security token
func (abc *AddressBookCollection) Delete(c *gin.Context) {
	f := &model.AddressBookCollection{}
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
	ex := service.AllService.AddressBookService.CollectionInfoById(f.Id)
	if ex.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
		return
	}
	err := service.AllService.AddressBookService.DeleteCollection(ex)
	if err == nil {
		response.Success(c, nil)
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
}
