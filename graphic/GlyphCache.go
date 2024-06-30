package graphic

import (
	"container/list"
	"image"
)

// GlyphCache 缓存结构
type GlyphCache struct {
	cache      map[rune]image.Image
	order      *list.List
	maxEntries int
}

func NewGlyphCache(maxEntries int) *GlyphCache {
	return &GlyphCache{
		cache:      make(map[rune]image.Image),
		order:      list.New(),
		maxEntries: maxEntries,
	}
}

func (gc *GlyphCache) Get(r rune) (image.Image, bool) {
	img, ok := gc.cache[r]
	return img, ok
}

func (gc *GlyphCache) Add(r rune, img image.Image) {
	if gc.order.Len() >= gc.maxEntries {
		// 删除最早使用的元素
		front := gc.order.Front()
		if front != nil {
			delete(gc.cache, front.Value.(rune))
			gc.order.Remove(front)
		}
	}
	gc.cache[r] = img
	gc.order.PushBack(r)
}
