package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"rustdesk-api/config"
	"rustdesk-api/global"
	"rustdesk-api/http/controller/api"
	"rustdesk-api/http/middleware"
	"rustdesk-api/lib/jwt"
	"rustdesk-api/model"
	"rustdesk-api/service"
)

func setupTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup in-memory SQLite DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	global.DB = db

	db.AutoMigrate(
		&model.User{},
		&model.UserToken{},
		&model.Peer{},
		&model.AuditConn{},
		&model.AuditAlarm{},
		&model.DeviceGroup{},
		&model.AddressBookCollection{},
		&model.AddressBook{},
	)

	// Setup Services
	jwtKey := "test-secret-key-that-is-long-enough-32bytes!!"
	global.Config.Jwt.Key = jwtKey
	j := jwt.NewJwt(jwtKey, 24*time.Hour)
	global.Jwt = j

	service.New(&config.Config{}, db, nil, j, nil)

	r := gin.New()
	au := &api.Audit{}
	pe := &api.Peer{}

	// Register routes
	r.POST("/api/audit/conn", au.AuditConn)
	r.POST("/api/audit/alarm", au.AuditAlarm)

	auth := r.Group("/api")
	auth.Use(middleware.RustAuth())
	auth.GET("/audit/conn/active", au.AuditConnActive)
	auth.PUT("/audit", au.UpdateAuditNote)
	auth.POST("/devices/deploy", pe.Deploy)
	auth.POST("/devices/cli", pe.Cli)

	return r, db
}

func TestAuditWorkflow(t *testing.T) {
	r, db := setupTestRouter()

	// 1. Create a test user and their token
	user := &model.User{
		Username: "testuser",
		Status:   model.COMMON_STATUS_ENABLE,
	}
	db.Create(user)

	token := global.Jwt.GenerateToken(user.Id)
	ut := &model.UserToken{
		UserId:    user.Id,
		Token:     token,
		ExpiredAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	db.Create(ut)

	// Create peers associated with the user
	peer1 := &model.Peer{
		Id:     "peer-src",
		UserId: user.Id,
	}
	peer2 := &model.Peer{
		Id:     "peer-dst",
		UserId: user.Id,
	}
	db.Create(peer1)
	db.Create(peer2)

	// 2. Test AuditConn (action = new) -> generates GUID
	connReqBody := map[string]interface{}{
		"action":     "new",
		"conn_id":    12345,
		"id":         "peer-dst",
		"peer":       []string{"peer-src", "Test Source"},
		"ip":         "127.0.0.1",
		"session_id": 99.0,
		"type":       1,
		"uuid":       "some-uuid",
	}
	bodyBytes, _ := json.Marshal(connReqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/audit/conn", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var connResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &connResp)
	guid, ok := connResp["data"].(string)
	if !ok || guid == "" {
		t.Fatalf("Expected returned GUID in data field, got response: %s", w.Body.String())
	}

	// 3. Test AuditConnActive
	// 3a. Tim thay (found, correct user)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/audit/conn/active?id=peer-dst&session_id=99&conn_type=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	var returnedGuid string
	json.Unmarshal(w.Body.Bytes(), &returnedGuid)
	if returnedGuid != guid {
		t.Fatalf("Expected returned GUID %s, got %s", guid, returnedGuid)
	}

	// 3b. Khong tim thay (not found)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/audit/conn/active?id=peer-dst&session_id=999&conn_type=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	// 3c. Sai user (wrong user - peer owned by someone else)
	// Create another user
	otherUser := &model.User{
		Username: "otheruser",
		Status:   model.COMMON_STATUS_ENABLE,
	}
	db.Create(otherUser)
	otherToken := global.Jwt.GenerateToken(otherUser.Id)
	otherUt := &model.UserToken{
		UserId:    otherUser.Id,
		Token:     otherToken,
		ExpiredAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	db.Create(otherUt)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/audit/conn/active?id=peer-dst&session_id=99&conn_type=1", nil)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status 403 (Forbidden), got %d", w.Code)
	}

	// 3d. Token khong hop le (invalid token)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/audit/conn/active?id=peer-dst&session_id=99&conn_type=1", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", w.Code)
	}

	// 4. Test UpdateAuditNote
	// 4a. Cap nhat thanh cong (success)
	noteReqBody := map[string]interface{}{
		"guid": guid,
		"note": "Successful connection test note",
	}
	noteBytes, _ := json.Marshal(noteReqBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/audit", bytes.NewBuffer(noteBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify note was updated in DB
	var dbConn model.AuditConn
	db.Where("guid = ?", guid).First(&dbConn)
	if dbConn.Note != "Successful connection test note" {
		t.Fatalf("Expected Note in DB to be updated, got '%s'", dbConn.Note)
	}

	// 4b. GUID sai (wrong GUID)
	wrongNoteReqBody := map[string]interface{}{
		"guid": "wrong-guid",
		"note": "Some note",
	}
	wrongNoteBytes, _ := json.Marshal(wrongNoteReqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/audit", bytes.NewBuffer(wrongNoteBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	// 4c. Sai user (wrong user)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/audit", bytes.NewBuffer(noteBytes))
	req.Header.Set("Authorization", "Bearer "+otherToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status 403, got %d", w.Code)
	}

	// 5. Test AuditAlarm & rate limiting
	alarmReqBody := map[string]interface{}{
		"id":   "peer-src",
		"uuid": "base64-device-uuid",
		"typ":  1,
		"info": "Attempted illegal login",
	}
	alarmBytes, _ := json.Marshal(alarmReqBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/audit/alarm", bytes.NewBuffer(alarmBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Try sending again immediately -> should trigger rate limiting (429)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/audit/alarm", bytes.NewBuffer(alarmBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429 (Too Many Requests), got %d", w.Code)
	}

	// 6. Test Deploy Endpoint (/api/devices/deploy)
	// 6a. Deploy new peer
	deployReqBody := map[string]interface{}{
		"id":   "new-peer-deploy-id",
		"uuid": "new-peer-uuid-b64",
		"pk":   "new-peer-pk-b64",
	}
	deployBytes, _ := json.Marshal(deployReqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/devices/deploy", bytes.NewBuffer(deployBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected deploy status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	var deployResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &deployResp)
	if deployResp["result"] != "OK" {
		t.Fatalf("Expected deploy result to be 'OK', got '%v'", deployResp["result"])
	}

	// Verify peer created in DB
	var newPe model.Peer
	db.Where("id = ?", "new-peer-deploy-id").First(&newPe)
	if newPe.RowId == 0 || newPe.Uuid != "new-peer-uuid-b64" || newPe.Pk != "new-peer-pk-b64" || newPe.UserId != user.Id {
		t.Fatalf("Peer not properly created/deployed in DB: %+v", newPe)
	}

	// 6b. Deploy existing peer (owned by current user) -> updates pk/uuid
	deployReqBody2 := map[string]interface{}{
		"id":   "new-peer-deploy-id",
		"uuid": "updated-peer-uuid-b64",
		"pk":   "updated-peer-pk-b64",
	}
	deployBytes2, _ := json.Marshal(deployReqBody2)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/devices/deploy", bytes.NewBuffer(deployBytes2))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected deploy status 200, got %d", w.Code)
	}
	db.Where("id = ?", "new-peer-deploy-id").First(&newPe)
	if newPe.Uuid != "updated-peer-uuid-b64" || newPe.Pk != "updated-peer-pk-b64" {
		t.Fatalf("Peer not properly updated during deploy: %+v", newPe)
	}

	// 6c. Deploy existing peer owned by another user -> result: ID_TAKEN
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/devices/deploy", bytes.NewBuffer(deployBytes2))
	req.Header.Set("Authorization", "Bearer "+otherToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected deploy status 200, got %d", w.Code)
	}
	var deployResp2 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &deployResp2)
	if deployResp2["result"] != "ID_TAKEN" {
		t.Fatalf("Expected deploy result to be 'ID_TAKEN', got '%v'", deployResp2["result"])
	}

	// 7. Test Cli Endpoint (/api/devices/cli)
	// Create a device group
	deviceGroup := &model.DeviceGroup{
		Name: "test-device-group",
	}
	db.Create(deviceGroup)

	cliReqBody := map[string]interface{}{
		"id":                "new-peer-deploy-id",
		"uuid":              "updated-peer-uuid-b64",
		"device_group_name": "test-device-group",
		"note":              "new remark/note",
		"device_username":   "new-device-username",
		"device_name":       "new-device-name",
		"address_book_name": "my-addr-book",
		"address_book_tag":  "tagA,tagB",
		"address_book_note": "address book note value",
	}
	cliBytes, _ := json.Marshal(cliReqBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/devices/cli", bytes.NewBuffer(cliBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected cli status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify updates on Peer in DB
	db.Where("id = ?", "new-peer-deploy-id").First(&newPe)
	if newPe.GroupId != deviceGroup.Id || newPe.Alias != "new remark/note" || newPe.Username != "new-device-username" || newPe.Hostname != "new-device-name" {
		t.Fatalf("Peer fields not updated by CLI: %+v", newPe)
	}

	// Verify Address Book entry created in DB
	var dbAb model.AddressBook
	db.Where("id = ? AND user_id = ?", "new-peer-deploy-id", user.Id).First(&dbAb)
	if dbAb.RowId == 0 || dbAb.Username != "new-device-username" || dbAb.Hostname != "new-device-name" || dbAb.LoginName != "address book note value" {
		t.Fatalf("Address Book entry not properly created by CLI: %+v", dbAb)
	}

	// Test invalid device ID (not found) -> 404
	cliReqBodyWrongId := map[string]interface{}{
		"id":   "non-existent-id",
		"uuid": "some-uuid",
	}
	cliWrongIdBytes, _ := json.Marshal(cliReqBodyWrongId)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/devices/cli", bytes.NewBuffer(cliWrongIdBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected CLI status 404, got %d", w.Code)
	}

	// Test unauthorized access (other user trying to manage peer) -> 403
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/devices/cli", bytes.NewBuffer(cliBytes))
	req.Header.Set("Authorization", "Bearer "+otherToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected CLI status 403, got %d", w.Code)
	}
}
