/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"strconv"
	"sync/atomic"
)

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
