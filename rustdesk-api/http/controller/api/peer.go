package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"rustdesk-api/global"
	requstform "rustdesk-api/http/request/api"
	"rustdesk-api/http/response"
	"rustdesk-api/model"
	"rustdesk-api/model/custom_types"
	"rustdesk-api/service"
	"strings"
)

type Peer struct {
}

// SysInfo
// @Tags System
// @Summary Submit system information
// @Description Submit system information
// @Accept  json
// @Produce  json
// @Param body body requstform.PeerForm true "System information form"
// @Success 200 {string} string "SYSINFO_UPDATED,ID_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo [post]
func (p *Peer) SysInfo(c *gin.Context) {
	f := &requstform.PeerForm{}
	err := c.ShouldBindBodyWith(f, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	fpe := f.ToPeer()
	pe := service.AllService.PeerService.FindById(f.Id)
	if pe.RowId == 0 {
		pe = f.ToPeer()
		pe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(pe.Uuid, pe.Id)
		err = service.AllService.PeerService.Create(pe)
		if err != nil {
			response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
			return
		}
	} else {
		if pe.UserId == 0 {
			pe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(pe.Uuid, pe.Id)
		}
		fpe.RowId = pe.RowId
		fpe.UserId = pe.UserId
		err = service.AllService.PeerService.Update(fpe)
		if err != nil {
			response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
			return
		}
	}
	//SYSINFO_UPDATED uploaded successfully
	//ID_NOT_FOUND The next heartbeat will be uploaded
	//direct response text
	c.String(http.StatusOK, "SYSINFO_UPDATED")
}

// SysInfoVer
// @Tags System
// @Summary Get system version information
// @Description Get system version information
// @Accept  json
// @Produce  json
// @Success 200 {string} string ""
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo_ver [post]
func (p *Peer) SysInfoVer(c *gin.Context) {
	//Read resources/version file
	v := service.AllService.AppService.GetAppVersion()
	// Add the startup time to facilitate the client to upload information
	v = fmt.Sprintf("%s\n%s", v, service.AllService.AppService.GetStartTime())
	c.String(http.StatusOK, v)
}

type DeployForm struct {
	Id   string `json:"id" binding:"required"`
	Uuid string `json:"uuid" binding:"required"`
	Pk   string `json:"pk" binding:"required"`
}

func (p *Peer) Deploy(c *gin.Context) {
	var form DeployForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "INVALID_INPUT"})
		return
	}

	currentUser := service.AllService.UserService.CurUser(c)
	if currentUser == nil || currentUser.Id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa xác thực"})
		return
	}

	pe := service.AllService.PeerService.FindById(form.Id)
	if pe.RowId > 0 {
		if pe.UserId != 0 && pe.UserId != currentUser.Id {
			c.JSON(http.StatusOK, gin.H{"result": "ID_TAKEN"})
			return
		}
		pe.Uuid = form.Uuid
		pe.Pk = form.Pk
		pe.UserId = currentUser.Id
		err := service.AllService.PeerService.Update(pe)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"result": err.Error()})
			return
		}
	} else {
		newPeer := &model.Peer{
			Id:     form.Id,
			Uuid:   form.Uuid,
			Pk:     form.Pk,
			UserId: currentUser.Id,
		}
		err := service.AllService.PeerService.Create(newPeer)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"result": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "OK",
	})
}

type DeviceCliForm struct {
	Id                  string `json:"id" binding:"required"`
	Uuid                string `json:"uuid" binding:"required"`
	UserName            string `json:"user_name"`
	StrategyName        string `json:"strategy_name"`
	AddressBookName     string `json:"address_book_name"`
	AddressBookTag      string `json:"address_book_tag"`
	AddressBookAlias    string `json:"address_book_alias"`
	AddressBookPassword string `json:"address_book_password"`
	AddressBookNote     string `json:"address_book_note"`
	DeviceGroupName     string `json:"device_group_name"`
	Note                string `json:"note"`
	DeviceUsername      string `json:"device_username"`
	DeviceName          string `json:"device_name"`
}

func (p *Peer) Cli(c *gin.Context) {
	var form DeviceCliForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser := service.AllService.UserService.CurUser(c)
	if currentUser == nil || currentUser.Id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa xác thực"})
		return
	}

	pe := service.AllService.PeerService.FindById(form.Id)
	if pe.RowId == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}

	isAdmin := currentUser.IsAdmin != nil && *currentUser.IsAdmin
	if !isAdmin && pe.UserId != currentUser.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không có quyền quản lý thiết bị này"})
		return
	}

	if form.UserName != "" {
		var targetUser model.User
		err := global.DB.Where("username = ?", form.UserName).First(&targetUser).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy người dùng có tên đăng nhập '" + form.UserName + "'"})
			return
		}
		pe.UserId = targetUser.Id
	}

	if form.DeviceGroupName != "" {
		var deviceGroup model.DeviceGroup
		err := global.DB.Where("name = ?", form.DeviceGroupName).First(&deviceGroup).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy nhóm thiết bị '" + form.DeviceGroupName + "'"})
			return
		}
		pe.GroupId = deviceGroup.Id
	}

	if form.AddressBookName != "" {
		var collection model.AddressBookCollection
		err := global.DB.Where("name = ? AND user_id = ?", form.AddressBookName, currentUser.Id).First(&collection).Error
		if err != nil {
			collection = model.AddressBookCollection{
				UserId: currentUser.Id,
				Name:   form.AddressBookName,
			}
			err = global.DB.Create(&collection).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo bộ sưu tập sổ địa chỉ: " + err.Error()})
				return
			}
		}

		ab := service.AllService.AddressBookService.InfoByUserIdAndIdAndCid(currentUser.Id, form.Id, collection.Id)
		var tags custom_types.AutoJson
		if form.AddressBookTag != "" {
			tagsList := strings.Split(form.AddressBookTag, ",")
			tagsBytes, _ := json.Marshal(tagsList)
			tags = custom_types.AutoJson(tagsBytes)
		} else {
			tags = custom_types.AutoJson([]byte("[]"))
		}

		if ab.RowId > 0 {
			ab.Alias = form.AddressBookAlias
			ab.Password = form.AddressBookPassword
			ab.Hash = form.AddressBookPassword
			ab.Tags = tags
			if form.AddressBookNote != "" {
				ab.LoginName = form.AddressBookNote
			}
			if form.DeviceUsername != "" {
				ab.Username = form.DeviceUsername
			}
			if form.DeviceName != "" {
				ab.Hostname = form.DeviceName
			}
			global.DB.Save(ab)
		} else {
			newAb := &model.AddressBook{
				Id:           form.Id,
				UserId:       currentUser.Id,
				CollectionId: collection.Id,
				Alias:        form.AddressBookAlias,
				Password:     form.AddressBookPassword,
				Hash:         form.AddressBookPassword,
				Tags:         tags,
				Username:     form.DeviceUsername,
				Hostname:     form.DeviceName,
			}
			if form.AddressBookNote != "" {
				newAb.LoginName = form.AddressBookNote
			}
			err = service.AllService.AddressBookService.AddAddressBook(newPeerAddressBook(newAb, pe))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo mục sổ địa chỉ: " + err.Error()})
				return
			}
		}
	}

	if form.Note != "" {
		pe.Alias = form.Note
	}

	if form.DeviceUsername != "" {
		pe.Username = form.DeviceUsername
	}

	if form.DeviceName != "" {
		pe.Hostname = form.DeviceName
	}

	err := service.AllService.PeerService.Update(pe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, "")
}

func newPeerAddressBook(ab *model.AddressBook, pe *model.Peer) *model.AddressBook {
	if ab.Username == "" {
		ab.Username = pe.Username
	}
	if ab.Hostname == "" {
		ab.Hostname = pe.Hostname
	}
	ab.Platform = service.AllService.AddressBookService.PlatformFromOs(pe.Os)
	return ab
}
