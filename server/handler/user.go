package handler

import (
	"strconv"

	"daidai-panel/database"
	"daidai-panel/middleware"
	"daidai-panel/model"
	"daidai-panel/pkg/crypto"
	"daidai-panel/pkg/response"
	"daidai-panel/pkg/validator"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) List(c *gin.Context) {
	var users []model.User
	database.DB.Order("created_at ASC").Find(&users)

	data := make([]map[string]interface{}, len(users))
	for i, u := range users {
		data[i] = u.ToDict()
	}

	response.Success(c, gin.H{"data": data})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if !validator.ValidateUsername(req.Username) {
		response.BadRequest(c, "用户名格式不正确")
		return
	}
	if !validator.ValidatePassword(req.Password) {
		response.BadRequest(c, "密码长度需 6-128 位")
		return
	}
	if req.Role == "" {
		req.Role = "operator"
	}
	if req.Role != "admin" && req.Role != "operator" && req.Role != "viewer" {
		response.BadRequest(c, "角色无效，可选 admin/operator/viewer")
		return
	}

	var existing model.User
	if database.DB.Where("username = ?", req.Username).First(&existing).Error == nil {
		response.BadRequest(c, "用户名已存在")
		return
	}

	hashed, err := crypto.HashPassword(req.Password)
	if err != nil {
		response.InternalError(c, "密码加密失败")
		return
	}

	user := model.User{
		Username: req.Username,
		Password: hashed,
		Role:     req.Role,
		Enabled:  true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		response.InternalError(c, "创建用户失败")
		return
	}

	response.Created(c, gin.H{"message": "创建成功", "data": user.ToDict()})
}

func (h *UserHandler) Update(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	var req struct {
		Role    string `json:"role"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Role != "" {
		if req.Role != "admin" && req.Role != "operator" && req.Role != "viewer" {
			response.BadRequest(c, "角色无效")
			return
		}
		updates["role"] = req.Role
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		database.DB.Model(&user).Updates(updates)
	}

	database.DB.First(&user, userID)
	response.Success(c, gin.H{"message": "更新成功", "data": user.ToDict()})
}

func (h *UserHandler) Delete(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	currentUser := c.GetString("username")
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	if user.Username == currentUser {
		response.BadRequest(c, "不能删除自己")
		return
	}

	database.DB.Delete(&user)
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if !validator.ValidatePassword(req.Password) {
		response.BadRequest(c, "密码长度需 6-128 位")
		return
	}

	hashed, err := crypto.HashPassword(req.Password)
	if err != nil {
		response.InternalError(c, "密码加密失败")
		return
	}

	database.DB.Model(&user).Update("password", hashed)
	response.Success(c, gin.H{"message": "密码重置成功"})
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users", middleware.JWTAuth(), middleware.RequireAdmin())
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.PUT("/:id/reset-password", h.ResetPassword)
	}
}
