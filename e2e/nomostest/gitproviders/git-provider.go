// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitproviders

import (
	"kpt.dev/configsync/e2e"
	"kpt.dev/configsync/e2e/nomostest/portforwarder"
	"kpt.dev/configsync/e2e/nomostest/testing"
)

const (
	// GitUser is the user for all Git providers.
	GitUser = "config-sync-ci-bot"
)

// GitProvider is an interface for the remote Git providers.
type GitProvider interface {
	Type() string

	// RemoteURL returns remote URL of the repository.
	// It is used to set the url for the remote origin using `git remote add origin <REMOTE_URL>.
	// For the testing git-server, RemoteURL uses localhost and forwarded port, while SyncURL uses the DNS.
	// For other git providers, RemoteURL should be the same as SyncURL.
	// name refers to the repo name in the format of <NAMESPACE>/<NAME> of RootSync|RepoSync.
	RemoteURL(name string, opts ...RemoteURLOpt) (string, error)

	// SyncURL returns the git repository URL for Config Sync to sync from.
	// name refers to the repo name in the format of <NAMESPACE>/<NAME> of RootSync|RepoSync.
	SyncURL(name string) string
	CreateRepository(name string) (string, error)
	DeleteRepositories(names ...string) error
	DeleteObsoleteRepos() error
}

// GitProviderOpt is an optional parameter for instantiating a new GitProvider
type GitProviderOpt func(opts *GitProviderOpts)

// GitProviderOpts is the set of optional parameters for instantiating a new GitProvider
type GitProviderOpts struct {
	portForwarder *portforwarder.PortForwarder
}

// WithPortForwarder provides a PortForwarder for the GitProvider to use.
// Required for LocalProvider in order to establish a PortForwarder to the in-cluster
// git server.
func WithPortForwarder(portForwarder *portforwarder.PortForwarder) GitProviderOpt {
	return func(opts *GitProviderOpts) {
		opts.portForwarder = portForwarder
	}
}

// RemoteURLOpt is an optional parameter for calling RemoteURL
type RemoteURLOpt func(opts *RemoteURLOpts)

// RemoteURLOpts is the set of optional parameters for calling RemoteURL
type RemoteURLOpts struct {
	localPort int
}

// WithLocalPort provides a localPort to use when constructing the RemoteURL.
// This port is used by the local provider instead of getting the LocalPort from
// the PortForwarder. This can be useful in certain scenarios to avoid deadlock.
func WithLocalPort(localPort int) RemoteURLOpt {
	return func(opts *RemoteURLOpts) {
		opts.localPort = localPort
	}
}

// NewGitProvider creates a GitProvider for the specific provider type.
func NewGitProvider(t testing.NTB, provider string, opts ...GitProviderOpt) GitProvider {
	options := GitProviderOpts{}
	for _, opt := range opts {
		opt(&options)
	}
	switch provider {
	case e2e.Bitbucket:
		client, err := newBitbucketClient()
		if err != nil {
			t.Fatal(err)
		}
		return client
	case e2e.GitLab:
		client, err := newGitlabClient()
		if err != nil {
			t.Fatal((err))
		}
		return client
	default:
		client, err := newLocalProvider(options)
		if err != nil {
			t.Fatal(err)
		}
		return client
	}
}
