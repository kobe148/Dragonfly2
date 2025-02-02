/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dfpath

import (
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		expect  func(t *testing.T, options []Option)
	}{
		{
			name:    "new dfpath failed",
			options: []Option{WithLogDir("")},
			expect: func(t *testing.T, options []Option) {
				assert := assert.New(t)
				_, err := New(options...)
				assert.Error(err)
			},
		},
		{
			name: "new dfpath",
			expect: func(t *testing.T, options []Option) {
				assert := assert.New(t)
				cache.Once = sync.Once{}
				cache.err = &multierror.Error{}
				d, err := New(options...)
				assert.NoError(err)
				assert.Equal(d.WorkHome(), DefaultWorkHome)
				assert.Equal(d.CacheDir(), "")
				assert.Equal(d.LogDir(), DefaultLogDir)
				assert.Equal(d.DataDir(), "")
				assert.Equal(d.PluginDir(), DefaultPluginDir)
				assert.Equal(d.DaemonSockPath(), DefaultDownloadUnixSocketPath)
			},
		},
		{
			name:    "new dfpath by workHome",
			options: []Option{WithWorkHome("foo")},
			expect: func(t *testing.T, options []Option) {
				assert := assert.New(t)
				cache.Once = sync.Once{}
				cache.err = &multierror.Error{}
				d, err := New(options...)
				assert.NoError(err)
				assert.Equal(d.WorkHome(), "foo")
				assert.Equal(d.CacheDir(), "")
				assert.Equal(d.LogDir(), DefaultLogDir)
				assert.Equal(d.DataDir(), "")
				assert.Equal(d.PluginDir(), DefaultPluginDir)
			},
		},
		{
			name:    "new dfpath by cacheDir",
			options: []Option{WithCacheDir("foo")},
			expect: func(t *testing.T, options []Option) {
				assert := assert.New(t)
				cache.Once = sync.Once{}
				cache.err = &multierror.Error{}
				d, err := New(options...)
				assert.NoError(err)
				assert.Equal(d.WorkHome(), DefaultWorkHome)
				assert.Equal(d.CacheDir(), "foo")
				assert.Equal(d.LogDir(), DefaultLogDir)
				assert.Equal(d.DataDir(), "")
				assert.Equal(d.PluginDir(), DefaultPluginDir)
				assert.Equal(d.DaemonSockPath(), DefaultDownloadUnixSocketPath)
			},
		},
		{
			name:    "new dfpath by logDir",
			options: []Option{WithLogDir("foo")},
			expect: func(t *testing.T, options []Option) {
				assert := assert.New(t)
				cache.Once = sync.Once{}
				cache.err = &multierror.Error{}
				d, err := New(options...)
				assert.NoError(err)
				assert.Equal(d.WorkHome(), DefaultWorkHome)
				assert.Equal(d.CacheDir(), "")
				assert.Equal(d.LogDir(), "foo")
				assert.Equal(d.DataDir(), "")
				assert.Equal(d.PluginDir(), DefaultPluginDir)
				assert.Equal(d.DaemonSockPath(), DefaultDownloadUnixSocketPath)
			},
		},
		{
			name:    "new dfpath by pluginDir",
			options: []Option{WithPluginDir("foo")},
			expect: func(t *testing.T, options []Option) {
				assert := assert.New(t)
				cache.Once = sync.Once{}
				cache.err = &multierror.Error{}
				d, err := New(options...)
				assert.NoError(err)
				assert.Equal(d.WorkHome(), DefaultWorkHome)
				assert.Equal(d.CacheDir(), "")
				assert.Equal(d.LogDir(), DefaultLogDir)
				assert.Equal(d.DataDir(), "")
				assert.Equal(d.PluginDir(), "foo")
				assert.Equal(d.DaemonSockPath(), DefaultDownloadUnixSocketPath)
			},
		},
		{
			name:    "new dfpath by daemonSockPath",
			options: []Option{WithDownloadUnixSocketPath("foo")},
			expect: func(t *testing.T, options []Option) {
				assert := assert.New(t)
				cache.Once = sync.Once{}
				cache.err = &multierror.Error{}
				d, err := New(options...)
				assert.NoError(err)
				assert.Equal(d.WorkHome(), DefaultWorkHome)
				assert.Equal(d.CacheDir(), "")
				assert.Equal(d.LogDir(), DefaultLogDir)
				assert.Equal(d.DataDir(), "")
				assert.Equal(d.PluginDir(), DefaultPluginDir)
				assert.Equal(d.DaemonSockPath(), "foo")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expect(t, tc.options)
		})
	}
}
