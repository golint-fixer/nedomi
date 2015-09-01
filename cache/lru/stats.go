package lru

// This file contains the LRUCache's implementation of the CacheStats interface.

import (
	"fmt"

	"github.com/ironsmile/nedomi/types"
)

//!TODO: move these stats - they are not specific to the LRU cache

// TieredCacheStats is used by the LRUCache to implement the CacheStats interface.
type TieredCacheStats struct {
	id       string
	hits     uint64
	requests uint64
	size     types.BytesSize
	objects  uint64
}

// CacheHitPrc implements part of CacheStats interface
func (lcs *TieredCacheStats) CacheHitPrc() string {
	if lcs.requests == 0 {
		return ""
	}
	return fmt.Sprintf("%.f%%", (float32(lcs.Hits())/float32(lcs.Requests()))*100)
}

// ID implements part of CacheStats interface
func (lcs *TieredCacheStats) ID() string {
	return lcs.id
}

// Hits implements part of CacheStats interface
func (lcs *TieredCacheStats) Hits() uint64 {
	return lcs.hits
}

// Size implements part of CacheStats interface
func (lcs *TieredCacheStats) Size() types.BytesSize {
	return lcs.size
}

// Objects implements part of CacheStats interface
func (lcs *TieredCacheStats) Objects() uint64 {
	return lcs.objects
}

// Requests implements part of CacheStats interface
func (lcs *TieredCacheStats) Requests() uint64 {
	return lcs.requests
}

// Stats implements part of types.CacheAlgorithm interface
func (tc *TieredLRUCache) Stats() types.CacheStats {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	var sum types.BytesSize
	var allObjects uint64

	for i := 0; i < cacheTiers; i++ {
		objects := types.BytesSize(tc.tiers[i].Len())
		sum += (tc.cfg.PartSize * objects)
		allObjects += uint64(objects)
	}

	return &TieredCacheStats{
		id:       tc.cfg.Path,
		hits:     tc.hits,
		requests: tc.requests,
		size:     sum,
		objects:  allObjects,
	}
}
