// The MIT License (MIT)
//
// Copyright (c) 2020 PostHog (part of Hiberly Inc)
//
// Copyright (c) 2016 Segment, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sysinfo

import (
	"runtime"
	"sync"
)

type OsInfo struct {
	Name    string `json:"$os"`
	Version string `json:"$os_version,omitempty"`
	Distro  string `json:"$os_distro,omitempty"`
}

type SysContext struct {
	OsInfo
	GoVersion string `json:"$go_version"`
}

var (
	cachedSystemContext     SysContext
	cachedSystemContextOnce sync.Once
	// cachedSysContextProps holds the pre-built properties from the cached system context.
	// This avoids creating a new Properties map on every ToProperties() call.
	cachedSysContextProps     map[string]any
	cachedSysContextPropsOnce sync.Once
)

// GetSystemContext returns the current process's OS and Go runtime info, computed
// once per process since neither changes during a run.
func GetSystemContext() SysContext {
	cachedSystemContextOnce.Do(func() {
		cachedSystemContext = SysContext{
			OsInfo:    getOSInfo(),
			GoVersion: runtime.Version(),
		}
	})
	return cachedSystemContext
}

func (ctx SysContext) ToProperties() map[string]any {
	cachedSysContextPropsOnce.Do(func() {
		sc := GetSystemContext()
		props := map[string]any{
			"$os":         sc.Name,
			"$go_version": sc.GoVersion,
		}
		if sc.Version != "" {
			props["$os_version"] = sc.Version
		}
		if sc.Distro != "" {
			props["$os_distro"] = sc.Distro
		}
		cachedSysContextProps = props
	})
	return cachedSysContextProps
}
