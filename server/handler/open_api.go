package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"daidai-panel/database"
	"daidai-panel/middleware"
	"daidai-panel/model"
	"daidai-panel/pkg/crypto"
	"daidai-panel/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"daidai-panel/config"
)

type OpenAPIHandler struct{}

func NewOpenAPIHandler() *OpenAPIHandler {
	return &OpenAPIHandler{}
}

func generateRandomKey(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (h *OpenAPIHandler) List(c *gin.Context) {
	var apps []model.OpenApp
	database.DB.Order("created_at DESC").Find(&apps)

	data := make([]map[string]interface{}, len(apps))
	for i, a := range apps {
		data[i] = a.ToDict()
	}

	response.Success(c, gin.H{"data": data})
}

func (h *OpenAPIHandler) Create(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Scopes    string `json:"scopes"`
		RateLimit int    `json:"rate_limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.RateLimit <= 0 {
		req.RateLimit = 100
	}

	app := model.OpenApp{
		Name:      req.Name,
		AppKey:    generateRandomKey(16),
		AppSecret: generateRandomKey(32),
		Scopes:    req.Scopes,
		Enabled:   true,
		RateLimit: req.RateLimit,
	}

	if err := database.DB.Create(&app).Error; err != nil {
		response.InternalError(c, "创建应用失败")
		return
	}

	response.Created(c, gin.H{"message": "创建成功", "data": app.ToDictWithSecret()})
}

func (h *OpenAPIHandler) Update(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var app model.OpenApp
	if err := database.DB.First(&app, appID).Error; err != nil {
		response.NotFound(c, "应用不存在")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	allowed := map[string]bool{"name": true, "scopes": true, "rate_limit": true}
	updates := make(map[string]interface{})
	for k, v := range req {
		if allowed[k] {
			updates[k] = v
		}
	}

	if len(updates) > 0 {
		database.DB.Model(&app).Updates(updates)
	}

	database.DB.First(&app, appID)
	response.Success(c, gin.H{"message": "更新成功", "data": app.ToDict()})
}

func (h *OpenAPIHandler) Delete(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	database.DB.Where("id = ?", appID).Delete(&model.OpenApp{})
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *OpenAPIHandler) Enable(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	database.DB.Model(&model.OpenApp{}).Where("id = ?", appID).Update("enabled", true)
	response.Success(c, gin.H{"message": "已启用"})
}

func (h *OpenAPIHandler) Disable(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	database.DB.Model(&model.OpenApp{}).Where("id = ?", appID).Update("enabled", false)
	response.Success(c, gin.H{"message": "已禁用"})
}

func (h *OpenAPIHandler) ResetSecret(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var app model.OpenApp
	if err := database.DB.First(&app, appID).Error; err != nil {
		response.NotFound(c, "应用不存在")
		return
	}

	newSecret := generateRandomKey(32)
	database.DB.Model(&app).Update("app_secret", newSecret)
	app.AppSecret = newSecret

	response.Success(c, gin.H{"message": "密钥已重置", "data": app.ToDictWithSecret()})
}

func (h *OpenAPIHandler) ViewSecret(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入密码")
		return
	}

	username := c.GetString("username")
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}

	if !crypto.CheckPassword(req.Password, user.Password) {
		response.Unauthorized(c, "密码错误")
		return
	}

	var app model.OpenApp
	if err := database.DB.First(&app, appID).Error; err != nil {
		response.NotFound(c, "应用不存在")
		return
	}

	response.Success(c, gin.H{"data": gin.H{"app_secret": app.AppSecret}})
}

func (h *OpenAPIHandler) Token(c *gin.Context) {
	var req struct {
		AppKey    string `json:"app_key" binding:"required"`
		AppSecret string `json:"app_secret" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var app model.OpenApp
	if err := database.DB.Where("app_key = ?", req.AppKey).First(&app).Error; err != nil {
		response.Unauthorized(c, "凭证无效")
		return
	}

	if !app.Enabled {
		response.Forbidden(c, "应用已被禁用")
		return
	}

	if app.AppSecret != req.AppSecret {
		response.Unauthorized(c, "凭证无效")
		return
	}

	claims := &middleware.Claims{
		Username:  fmt.Sprintf("app:%s", app.AppKey),
		Role:      fmt.Sprintf("app:%s", app.Scopes),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.C.JWT.Secret))
	if err != nil {
		response.InternalError(c, "生成令牌失败")
		return
	}

	database.DB.Model(&app).Update("call_count", app.CallCount+1)

	response.Success(c, gin.H{
		"data": gin.H{
			"access_token": tokenStr,
			"token_type":   "Bearer",
			"expires_in":   86400,
		},
	})
}

func (h *OpenAPIHandler) CallLogs(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.ApiCallLog{}).Where("app_id = ?", appID)

	var total int64
	query.Count(&total)

	var logs []model.ApiCallLog
	query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	data := make([]map[string]interface{}, len(logs))
	for i, l := range logs {
		data[i] = l.ToDict()
	}

	response.Paginated(c, data, total, page, pageSize)
}

func (h *OpenAPIHandler) RegisterRoutes(r *gin.RouterGroup) {
	openapi := r.Group("/open-api")
	{
		openapi.POST("/token", h.Token)

		mgmt := openapi.Group("", middleware.JWTAuth(), middleware.RequireAdmin())
		{
			mgmt.GET("/apps", h.List)
			mgmt.POST("/apps", h.Create)
			mgmt.PUT("/apps/:id", h.Update)
			mgmt.DELETE("/apps/:id", h.Delete)
			mgmt.PUT("/apps/:id/enable", h.Enable)
			mgmt.PUT("/apps/:id/disable", h.Disable)
			mgmt.PUT("/apps/:id/reset-secret", h.ResetSecret)
			mgmt.POST("/apps/:id/view-secret", h.ViewSecret)
			mgmt.GET("/apps/:id/logs", h.CallLogs)
		}
	}
}
