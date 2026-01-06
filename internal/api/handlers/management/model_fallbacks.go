package management

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
)

// GetModelFallbacks returns the current model fallback configuration.
func (h *Handler) GetModelFallbacks(c *gin.Context) {
	if h == nil || h.cfg == nil {
		c.JSON(http.StatusOK, gin.H{
			"model-fallbacks":      []config.ModelFallback{},
			"model-fallback-depth": 3,
		})
		return
	}
	depth := 3
	if h.cfg.ModelFallbackDepth != nil {
		depth = *h.cfg.ModelFallbackDepth
	}
	c.JSON(http.StatusOK, gin.H{
		"model-fallbacks":      h.cfg.ModelFallbacks,
		"model-fallback-depth": depth,
	})
}

// PutModelFallbacks replaces the entire model fallback configuration.
func (h *Handler) PutModelFallbacks(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var body struct {
		ModelFallbacks     []config.ModelFallback `json:"model-fallbacks"`
		ModelFallbackDepth *int                   `json:"model-fallback-depth"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if h.cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not available"})
		return
	}

	h.cfg.ModelFallbacks = body.ModelFallbacks
	h.cfg.ModelFallbackDepth = body.ModelFallbackDepth
	h.cfg.SanitizeModelFallbacks()

	if !h.persist(c) {
		return
	}

	h.updateFallbackConfig()
}

// PostModelFallbacks adds a single model fallback to the configuration.
func (h *Handler) PostModelFallbacks(c *gin.Context) {
	var fb config.ModelFallback
	if err := c.ShouldBindJSON(&fb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if h.cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not available"})
		return
	}

	for _, existing := range h.cfg.ModelFallbacks {
		if strings.EqualFold(existing.From, fb.From) && strings.EqualFold(existing.To, fb.To) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "fallback already exists"})
			return
		}
	}

	h.cfg.ModelFallbacks = append(h.cfg.ModelFallbacks, fb)
	h.cfg.SanitizeModelFallbacks()

	if !h.persist(c) {
		return
	}

	h.updateFallbackConfig()
}

// DeleteModelFallbacks removes model fallbacks from the configuration.
// Supports removing by: index, from model, or (from, to) pair.
func (h *Handler) DeleteModelFallbacks(c *gin.Context) {
	if h.cfg == nil || len(h.cfg.ModelFallbacks) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	if idxStr := c.Query("index"); idxStr != "" {
		var idx int
		if _, err := fmt.Sscanf(idxStr, "%d", &idx); err == nil && idx >= 0 && idx < len(h.cfg.ModelFallbacks) {
			h.cfg.ModelFallbacks = append(h.cfg.ModelFallbacks[:idx], h.cfg.ModelFallbacks[idx+1:]...)
			h.cfg.SanitizeModelFallbacks()
			if h.persist(c) {
				h.updateFallbackConfig()
			}
			return
		}
	}

	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))

	if from != "" && to == "" {
		out := make([]config.ModelFallback, 0, len(h.cfg.ModelFallbacks))
		for _, fb := range h.cfg.ModelFallbacks {
			if !strings.EqualFold(fb.From, from) {
				out = append(out, fb)
			}
		}
		if len(out) != len(h.cfg.ModelFallbacks) {
			h.cfg.ModelFallbacks = out
			h.cfg.SanitizeModelFallbacks()
			if h.persist(c) {
				h.updateFallbackConfig()
			}
			return
		}
	}

	if from != "" && to != "" {
		out := make([]config.ModelFallback, 0, len(h.cfg.ModelFallbacks))
		removed := false
		for _, fb := range h.cfg.ModelFallbacks {
			if strings.EqualFold(fb.From, from) && strings.EqualFold(fb.To, to) {
				removed = true
				continue
			}
			out = append(out, fb)
		}
		if removed {
			h.cfg.ModelFallbacks = out
			h.cfg.SanitizeModelFallbacks()
			if h.persist(c) {
				h.updateFallbackConfig()
			}
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "fallback not found"})
}

// updateFallbackConfig updates the auth manager with the current fallback configuration.
func (h *Handler) updateFallbackConfig() {
	if h.authManager != nil {
		h.authManager.SetFallbackConfig(h.cfg)
	}
}

// GetAvailableModels returns a list of all configured/available models for autocomplete.
func (h *Handler) GetAvailableModels(c *gin.Context) {
	if h == nil || h.cfg == nil {
		c.JSON(http.StatusOK, gin.H{"models": []string{}})
		return
	}

	models := make(map[string]bool)

	for _, key := range h.cfg.GeminiKey {
		models[key.APIKey] = false
		for _, m := range key.Models {
			models[m.Name] = true
			if m.Alias != "" {
				models[m.Alias] = true
			}
		}
	}
	for _, key := range h.cfg.ClaudeKey {
		for _, m := range key.Models {
			models[m.Name] = true
			if m.Alias != "" {
				models[m.Alias] = true
			}
		}
	}
	for _, key := range h.cfg.CodexKey {
		for _, m := range key.Models {
			models[m.Name] = true
			if m.Alias != "" {
				models[m.Alias] = true
			}
		}
	}
	for _, key := range h.cfg.OpenAICompatibility {
		for _, m := range key.Models {
			models[m.Name] = true
			if m.Alias != "" {
				models[m.Alias] = true
			}
		}
	}
	for _, key := range h.cfg.VertexCompatAPIKey {
		for _, m := range key.Models {
			models[m.Name] = true
			if m.Alias != "" {
				models[m.Alias] = true
			}
		}
	}

	for _, mapping := range h.cfg.AmpCode.ModelMappings {
		models[mapping.From] = true
		models[mapping.To] = true
	}

	for _, fb := range h.cfg.ModelFallbacks {
		models[fb.From] = true
		models[fb.To] = true
	}

	commonModels := []string{
		"gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.0-flash-exp",
		"claude-sonnet-4", "claude-haiku-3-5", "claude-opus-4",
		"gpt-4o", "gpt-4o-mini", "gpt-4-turbo",
	}
	for _, m := range commonModels {
		models[m] = true
	}

	modelList := make([]string, 0, len(models))
	for m := range models {
		modelList = append(modelList, m)
	}
	for i := 0; i < len(modelList); i++ {
		for j := i + 1; j < len(modelList); j++ {
			if modelList[i] > modelList[j] {
				modelList[i], modelList[j] = modelList[j], modelList[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": modelList})
}
