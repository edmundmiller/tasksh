package planning

import (
	"sync"
)

// widthCache caches visual width calculations for frequently used strings
var widthCache = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

// cachedVisualWidth returns the visual width of a string, using cache when possible
func cachedVisualWidth(s string) int {
	// For short strings, don't bother with caching
	if len(s) < 10 {
		return visualWidth(s)
	}

	// Check cache
	widthCache.RLock()
	if width, ok := widthCache.m[s]; ok {
		widthCache.RUnlock()
		return width
	}
	widthCache.RUnlock()

	// Calculate and cache
	width := visualWidth(s)
	
	widthCache.Lock()
	// Limit cache size to prevent unbounded growth
	if len(widthCache.m) > 1000 {
		// Clear cache if it gets too large
		widthCache.m = make(map[string]int)
	}
	widthCache.m[s] = width
	widthCache.Unlock()

	return width
}

// Optimized version of getContentWidth that caches the result per model
type contentWidthCache struct {
	width       int
	cachedWidth int
	valid       bool
}

func (c *contentWidthCache) get(currentWidth int) int {
	if c.valid && c.width == currentWidth {
		return c.cachedWidth
	}
	
	// Calculate new width
	contentWidth := currentWidth - 4 // Leave some margin
	if contentWidth < 80 {
		contentWidth = 80 // Minimum width
	}
	
	// Update cache
	c.width = currentWidth
	c.cachedWidth = contentWidth
	c.valid = true
	
	return contentWidth
}