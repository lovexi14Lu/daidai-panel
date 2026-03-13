package handler

import (
	"fmt"
	"strconv"

	"daidai-panel/database"
	"daidai-panel/middleware"
	"daidai-panel/model"
	"daidai-panel/pkg/response"
	"daidai-panel/service"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct{}

func NewSubscriptionHandler() *SubscriptionHandler {
	return &SubscriptionHandler{}
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	keyword := c.Query("keyword")
	subType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.Subscription{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name ILIKE ? OR url ILIKE ?", like, like)
	}
	if subType != "" {
		query = query.Where("type = ?", subType)
	}

	var total int64
	query.Count(&total)

	var subs []model.Subscription
	query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&subs)

	data := make([]map[string]interface{}, len(subs))
	for i, s := range subs {
		data[i] = s.ToDict()
	}

	response.Paginated(c, data, total, page, pageSize)
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type"`
		URL         string `json:"url" binding:"required"`
		Branch      string `json:"branch"`
		Schedule    string `json:"schedule"`
		Whitelist   string `json:"whitelist"`
		Blacklist   string `json:"blacklist"`
		DependOn    string `json:"depend_on"`
		AutoAddTask bool   `json:"auto_add_task"`
		AutoDelTask bool   `json:"auto_del_task"`
		SaveDir     string `json:"save_dir"`
		SSHKeyID    *uint  `json:"ssh_key_id"`
		Alias       string `json:"alias"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Type == "" {
		req.Type = model.SubTypeGitRepo
	}

	sub := model.Subscription{
		Name:        req.Name,
		Type:        req.Type,
		URL:         req.URL,
		Branch:      req.Branch,
		Schedule:    req.Schedule,
		Whitelist:   req.Whitelist,
		Blacklist:   req.Blacklist,
		DependOn:    req.DependOn,
		AutoAddTask: req.AutoAddTask,
		AutoDelTask: req.AutoDelTask,
		Enabled:     true,
		SaveDir:     req.SaveDir,
		SSHKeyID:    req.SSHKeyID,
		Alias:       req.Alias,
	}

	if err := database.DB.Create(&sub).Error; err != nil {
		response.InternalError(c, "创建订阅失败")
		return
	}

	response.Created(c, gin.H{"message": "创建成功", "data": sub.ToDict()})
}

func (h *SubscriptionHandler) Update(c *gin.Context) {
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var sub model.Subscription
	if err := database.DB.First(&sub, subID).Error; err != nil {
		response.NotFound(c, "订阅不存在")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	allowed := map[string]bool{
		"name": true, "type": true, "url": true, "branch": true,
		"schedule": true, "whitelist": true, "blacklist": true,
		"depend_on": true, "auto_add_task": true, "auto_del_task": true,
		"save_dir": true, "ssh_key_id": true, "alias": true,
	}
	updates := make(map[string]interface{})
	for k, v := range req {
		if allowed[k] {
			updates[k] = v
		}
	}

	if len(updates) > 0 {
		database.DB.Model(&sub).Updates(updates)
	}

	database.DB.First(&sub, subID)
	response.Success(c, gin.H{"message": "更新成功", "data": sub.ToDict()})
}

func (h *SubscriptionHandler) Delete(c *gin.Context) {
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	database.DB.Where("id = ?", subID).Delete(&model.Subscription{})
	database.DB.Where("subscription_id = ?", subID).Delete(&model.SubLog{})
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *SubscriptionHandler) Enable(c *gin.Context) {
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var sub model.Subscription
	if err := database.DB.First(&sub, subID).Error; err != nil {
		response.NotFound(c, "订阅不存在")
		return
	}
	database.DB.Model(&sub).Update("enabled", true)
	sub.Enabled = true
	response.Success(c, gin.H{"message": "已启用", "data": sub.ToDict()})
}

func (h *SubscriptionHandler) Disable(c *gin.Context) {
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var sub model.Subscription
	if err := database.DB.First(&sub, subID).Error; err != nil {
		response.NotFound(c, "订阅不存在")
		return
	}
	database.DB.Model(&sub).Update("enabled", false)
	sub.Enabled = false
	response.Success(c, gin.H{"message": "已禁用", "data": sub.ToDict()})
}

func (h *SubscriptionHandler) Pull(c *gin.Context) {
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var sub model.Subscription
	if err := database.DB.First(&sub, subID).Error; err != nil {
		response.NotFound(c, "订阅不存在")
		return
	}

	go service.PullSubscription(&sub)

	response.Success(c, gin.H{"message": "拉取任务已启动"})
}

func (h *SubscriptionHandler) Logs(c *gin.Context) {
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.SubLog{}).Where("subscription_id = ?", subID)

	var total int64
	query.Count(&total)

	var logs []model.SubLog
	query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	data := make([]map[string]interface{}, len(logs))
	for i, l := range logs {
		data[i] = l.ToDict()
	}

	response.Paginated(c, data, total, page, pageSize)
}

func (h *SubscriptionHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result := database.DB.Where("id IN ?", req.IDs).Delete(&model.Subscription{})
	database.DB.Where("subscription_id IN ?", req.IDs).Delete(&model.SubLog{})

	response.Success(c, gin.H{
		"message": fmt.Sprintf("已删除 %d 个订阅", result.RowsAffected),
	})
}

func (h *SubscriptionHandler) RegisterRoutes(r *gin.RouterGroup) {
	subs := r.Group("/subscriptions", middleware.JWTAuth())
	{
		subs.GET("", h.List)
		subs.POST("", h.Create)
		subs.PUT("/:id", h.Update)
		subs.DELETE("/:id", h.Delete)
		subs.PUT("/:id/enable", h.Enable)
		subs.PUT("/:id/disable", h.Disable)
		subs.PUT("/:id/pull", h.Pull)
		subs.GET("/:id/logs", h.Logs)
		subs.DELETE("/batch", h.BatchDelete)
	}
}
