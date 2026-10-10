/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	tokenAuthCacheTTL        = 5 * time.Minute
	tokenAuthCacheMaxEntries = 4096
)

type cachedTokenAuth struct {
	userId       string
	superAdmin   bool
	projectAdmin bool
	expiresAt    time.Time
}

// tokenAuthCache remembers successfully authenticated API keys, indexed by
// their digest so that no plain-text key is retained in memory.
type tokenAuthCache struct {
	mu      sync.RWMutex
	entries map[string]cachedTokenAuth
}

var apiKeyAuthCache = newTokenAuthCache()

func newTokenAuthCache() *tokenAuthCache {
	return &tokenAuthCache{
		entries: make(map[string]cachedTokenAuth),
	}
}

func (c *tokenAuthCache) get(apiKey string) (cachedTokenAuth, bool) {
	c.mu.RLock()
	entry, found := c.entries[apiKeyDigest(apiKey)]
	c.mu.RUnlock()
	if !found || !time.Now().Before(entry.expiresAt) {
		return cachedTokenAuth{}, false
	}
	return entry, true
}

func (c *tokenAuthCache) set(apiKey string, entry cachedTokenAuth) {
	now := time.Now()
	entry.expiresAt = now.Add(tokenAuthCacheTTL)

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= tokenAuthCacheMaxEntries {
		for k, e := range c.entries {
			if !now.Before(e.expiresAt) {
				delete(c.entries, k)
			}
		}
		// still full of valid entries: start over rather than grow unbounded
		if len(c.entries) >= tokenAuthCacheMaxEntries {
			clear(c.entries)
		}
	}
	c.entries[apiKeyDigest(apiKey)] = entry
}

// flush drops all cached API key authentications. It must be called whenever
// a token or its user changes (key renewal, expiration, removal, role change).
func (c *tokenAuthCache) flush() {
	c.mu.Lock()
	clear(c.entries)
	c.mu.Unlock()
}

// userResourcesGeneration is part of every cached user resources mapping key.
// Bumping it implicitly invalidates all of them at once (stale entries simply
// age out of the cache), which is required whenever project resources, project
// teams or team memberships change.
var userResourcesGeneration atomic.Uint64

func userResourcesCacheKey(userId string) string {
	return userId + "@" + strconv.FormatUint(userResourcesGeneration.Load(), 10)
}

func invalidateUserResources() {
	userResourcesGeneration.Add(1)
}
