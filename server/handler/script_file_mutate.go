package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"daidai-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ScriptHandler) SaveContent(c *gin.Context) {
	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	ext := strings.ToLower(filepath.Ext(req.Path))
	if ext != "" && !allowedExtensions[ext] {
		response.BadRequest(c, "不支持的文件类型")
		return
	}

	if len(req.Content) > maxUploadSize {
		response.BadRequest(c, "内容过大（最大 10MB）")
		return
	}

	full, err := safePath(req.Path, false)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	os.MkdirAll(filepath.Dir(full), 0755)
	if err := os.WriteFile(full, []byte(req.Content), 0644); err != nil {
		response.InternalError(c, "写入文件失败")
		return
	}

	newVersion := recordScriptVersion(req.Path, req.Content, req.Message)
	response.Success(c, gin.H{"message": "保存成功", "version": newVersion})
}

func (h *ScriptHandler) Upload(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "未选择文件")
		return
	}

	if header.Size > maxUploadSize {
		response.BadRequest(c, "文件过大（最大 10MB）")
		return
	}

	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" && !allowedExtensions[ext] {
		response.BadRequest(c, "不支持的文件类型")
		return
	}

	dir := c.PostForm("dir")
	targetPath := filename
	if dir != "" {
		targetPath = filepath.Join(dir, filename)
	}

	full, err := safePath(targetPath, false)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	os.MkdirAll(filepath.Dir(full), 0755)
	if err := c.SaveUploadedFile(header, full); err != nil {
		response.InternalError(c, "保存文件失败")
		return
	}

	response.Created(c, gin.H{"message": "上传成功", "path": targetPath})
}

func (h *ScriptHandler) Delete(c *gin.Context) {
	path := c.Query("path")
	fileType := c.DefaultQuery("type", "file")

	full, err := safePath(path, true)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if fileType == "directory" {
		os.RemoveAll(full)
	} else {
		os.Remove(full)
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *ScriptHandler) CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	full, err := safePath(req.Path, false)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := os.MkdirAll(full, 0755); err != nil {
		response.InternalError(c, "创建目录失败")
		return
	}

	response.Created(c, gin.H{"message": "创建成功"})
}

func (h *ScriptHandler) Rename(c *gin.Context) {
	var req struct {
		OldPath string `json:"old_path" binding:"required"`
		NewName string `json:"new_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if strings.ContainsAny(req.NewName, "/\\") {
		response.BadRequest(c, "新名称不能包含路径分隔符")
		return
	}

	full, err := safePath(req.OldPath, true)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	newFull := filepath.Join(filepath.Dir(full), req.NewName)
	if err := os.Rename(full, newFull); err != nil {
		response.InternalError(c, "重命名失败")
		return
	}

	response.Success(c, gin.H{"message": "重命名成功", "new_path": relPath(newFull)})
}

func (h *ScriptHandler) Move(c *gin.Context) {
	var req struct {
		SourcePath string `json:"source_path" binding:"required"`
		TargetDir  string `json:"target_dir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	srcFull, err := safePath(req.SourcePath, true)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	targetBase := scriptsDir()
	if req.TargetDir != "" {
		targetBase, err = safePath(req.TargetDir, true)
		if err != nil {
			response.BadRequest(c, "目标目录无效")
			return
		}
	}

	absTarget, _ := filepath.Abs(targetBase)
	absSrc, _ := filepath.Abs(srcFull)
	if strings.HasPrefix(absTarget, absSrc+string(filepath.Separator)) {
		response.BadRequest(c, "不能将目录移动到自身")
		return
	}

	destFull := filepath.Join(targetBase, filepath.Base(srcFull))
	if err := os.Rename(srcFull, destFull); err != nil {
		response.InternalError(c, "移动失败")
		return
	}

	response.Success(c, gin.H{"message": "移动成功", "new_path": relPath(destFull)})
}

func (h *ScriptHandler) Copy(c *gin.Context) {
	var req struct {
		SourcePath string `json:"source_path" binding:"required"`
		TargetDir  string `json:"target_dir"`
		NewName    string `json:"new_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	srcFull, err := safePath(req.SourcePath, true)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	targetBase := scriptsDir()
	if req.TargetDir != "" {
		targetBase, _ = safePath(req.TargetDir, false)
	}

	name := filepath.Base(srcFull)
	if req.NewName != "" {
		name = req.NewName
	}

	destFull := filepath.Join(targetBase, name)
	os.MkdirAll(targetBase, 0755)

	info, _ := os.Stat(srcFull)
	if info != nil && info.IsDir() {
		if err := copyDir(srcFull, destFull); err != nil {
			response.InternalError(c, "复制目录失败")
			return
		}
	} else {
		if err := copyFile(srcFull, destFull); err != nil {
			response.InternalError(c, "复制文件失败")
			return
		}
	}

	response.Created(c, gin.H{"message": "复制成功", "new_path": relPath(destFull)})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	os.MkdirAll(filepath.Dir(dst), 0755)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func (h *ScriptHandler) BatchDelete(c *gin.Context) {
	var req struct {
		Paths []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"paths" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	successCount := 0
	failedCount := 0
	failedItems := []string{}

	for _, item := range req.Paths {
		full, err := safePath(item.Path, true)
		if err != nil {
			failedCount++
			failedItems = append(failedItems, item.Path)
			continue
		}
		if item.Type == "directory" {
			err = os.RemoveAll(full)
		} else {
			err = os.Remove(full)
		}
		if err != nil {
			failedCount++
			failedItems = append(failedItems, item.Path)
		} else {
			successCount++
		}
	}

	response.Success(c, gin.H{
		"message":       fmt.Sprintf("删除完成: 成功 %d, 失败 %d", successCount, failedCount),
		"success_count": successCount,
		"failed_count":  failedCount,
		"failed_items":  failedItems,
	})
}
