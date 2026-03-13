package handler

import (
	"daidai-panel/database"
	"daidai-panel/middleware"
	"daidai-panel/model"
	"daidai-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

func (h *ConfigHandler) List(c *gin.Context) {
	var configs []model.SystemConfig
	database.DB.Order("key ASC").Find(&configs)

	data := make(map[string]interface{})
	for _, cfg := range configs {
		data[cfg.Key] = gin.H{
			"value":       cfg.Value,
			"description": cfg.Description,
			"updated_at":  cfg.UpdatedAt,
		}
	}

	response.Success(c, gin.H{"data": data})
}

func (h *ConfigHandler) Get(c *gin.Context) {
	key := c.Param("key")
	value := model.GetConfig(key, "")
	response.Success(c, gin.H{"data": gin.H{"key": key, "value": value}})
}

func (h *ConfigHandler) Set(c *gin.Context) {
	var req struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var cfg model.SystemConfig
	if err := database.DB.Where("`key` = ?", req.Key).First(&cfg).Error; err != nil {
		cfg = model.SystemConfig{
			Key:         req.Key,
			Value:       req.Value,
			Description: req.Description,
		}
		database.DB.Create(&cfg)
	} else {
		updates := map[string]interface{}{"value": req.Value}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		database.DB.Model(&cfg).Updates(updates)
	}

	response.Success(c, gin.H{"message": "配置已更新"})
}

func (h *ConfigHandler) BatchSet(c *gin.Context) {
	var req struct {
		Configs map[string]string `json:"configs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	for key, value := range req.Configs {
		model.SetConfig(key, value)
	}

	response.Success(c, gin.H{"message": "配置已更新"})
}

func (h *ConfigHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	database.DB.Where("`key` = ?", key).Delete(&model.SystemConfig{})
	response.Success(c, gin.H{"message": "配置已删除"})
}

func (h *ConfigHandler) RegisterRoutes(r *gin.RouterGroup) {
	cfgs := r.Group("/configs", middleware.JWTAuth(), middleware.RequireAdmin())
	{
		cfgs.GET("", h.List)
		cfgs.GET("/:key", h.Get)
		cfgs.POST("", h.Set)
		cfgs.PUT("/batch", h.BatchSet)
		cfgs.DELETE("/:key", h.Delete)
	}
}
