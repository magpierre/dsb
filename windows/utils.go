// Copyright 2025 Magnus Pierre
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package windows

import (
	"context"
	"time"
)

// createTimeoutContext creates a context with a configurable timeout for Delta Sharing API calls
// timeoutSeconds specifies the timeout duration in seconds (default: 60 seconds if <= 0)
func createTimeoutContext(timeoutSeconds int) (context.Context, context.CancelFunc) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60 // Default to 60 seconds
	}
	return context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
}
