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

package nomostest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kpt.dev/configsync/e2e/nomostest/gitproviders"
	"kpt.dev/configsync/e2e/nomostest/retry"
	"kpt.dev/configsync/e2e/nomostest/testing"
	"kpt.dev/configsync/e2e/nomostest/testkubeclient"
	"kpt.dev/configsync/e2e/nomostest/testlogger"
	"kpt.dev/configsync/pkg/core"
	"kpt.dev/configsync/pkg/importer/filesystem"
	"kpt.dev/configsync/pkg/syncer/reconcile"
	"kpt.dev/configsync/pkg/testing/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	// MainBranch is static as behavior when switching branches is never under
	// test.
	MainBranch = "main"

	// GitKeepFileName is a conventional name for an empty file you add to
	// directories in git when you want to make sure the directory is retained even
	// when all the other files are deleted.
	// Without this file, the directory may remain locally, but won't exist in git.
	GitKeepFileName = ".gitkeep"
)

// RepoType represents the type of the source repository.
type RepoType string

// RootRepo indicates the resources in the repository are cluster-scoped.
const RootRepo RepoType = "root"

// NamespaceRepo indicates the resources in the repository are namespace-scoped.
const NamespaceRepo RepoType = "namespace"

// Repository is a local git repository with a connection to a repository
// on the git-server for the test.
//
// We shell out for git commands as the git libraries are difficult to configure
// ssh for, and git-server requires ssh authentication.
type Repository struct {
	// Root is the location on the machine running the test at which the local
	// repository is stored.
	Root string
	// Format is the source format for parsing the repository (hierarchy or
	// unstructured).
	Format filesystem.SourceFormat

	T testing.NTB

	// Type refers to the type of the repository, i.e. if it is a root repo or a namespace repo.
	Type RepoType

	// SafetyNSPath is the path to the safety namespace yaml file.
	SafetyNSPath string

	// SafetyNSName is the name of the safety namespace.
	SafetyNSName string

	// SafetyClusterRolePath is the path to the safety namespace yaml file.
	SafetyClusterRolePath string

	// SafetyClusterRoleName is the name of the safety namespace.
	SafetyClusterRoleName string

	// RemoteRepoName is the name of the remote repository.
	// It is the same as Name for the testing git-server.
	// For other git providers, it appends a UUID to Name for uniqueness.
	RemoteRepoName string

	// GitProvider is the provider that hosts the Git repositories.
	GitProvider gitproviders.GitProvider

	// Scheme used for encoding and decoding objects.
	Scheme *runtime.Scheme

	// Logger for methods to use.
	Logger *testlogger.TestLogger
}

// NewRepository creates a remote repo on the git provider.
// Locally, it writes the repository to `tmpdir`/repos/`name`.
//
// The repo name is in the format of <NAMESPACE>/<NAME> of RootSync|RepoSync.
func NewRepository(nt *NT, repoType RepoType, nn types.NamespacedName, sourceFormat filesystem.SourceFormat) *Repository {
	nt.T.Helper()

	namespacedName := nn.String()
	safetyName := fmt.Sprintf("safety-%s", strings.ReplaceAll(namespacedName, "/", "-"))

	localDir := filepath.Join(nt.TmpDir, "repos", namespacedName)

	g := &Repository{
		Root:                  localDir,
		Format:                sourceFormat,
		T:                     nt.T,
		Type:                  repoType,
		SafetyNSName:          safetyName,
		SafetyNSPath:          fmt.Sprintf("acme/namespaces/%s/ns.yaml", safetyName),
		SafetyClusterRoleName: safetyName,
		SafetyClusterRolePath: fmt.Sprintf("acme/cluster/cluster-role-%s.yaml", safetyName),
		Scheme:                nt.Scheme,
		Logger:                nt.Logger,
		GitProvider:           nt.GitProvider,
	}

	repoName, err := nt.GitProvider.CreateRepository(namespacedName)
	// Add the repo to nt.RemoteRepositories immediately after it is created to reuse the repo.
	nt.RemoteRepositories[nn] = g
	if err != nil {
		nt.T.Fatal(err)
	}
	g.RemoteRepoName = repoName

	g.init(nt.gitPrivateKeyPath)
	g.initialCommit(sourceFormat)

	return g
}

// ReInit re-initializes the repo to the initial state.
func (g *Repository) ReInit(nt *NT, sourceFormat filesystem.SourceFormat) {
	nt.T.Helper()

	// Update test environment
	g.T = nt.T
	// Reset repo contents
	g.init(nt.gitPrivateKeyPath)
	g.initialCommit(sourceFormat)
}

func (g *Repository) gitCmd(command ...string) *exec.Cmd {
	// The -C flag executes git from repository root.
	// https://git-scm.com/docs/git#Documentation/git.txt--Cltpathgt
	args := []string{"git", "-C", g.Root}
	args = append(args, command...)
	g.Logger.Debugf("[repo %s] %s", path.Base(g.Root), strings.Join(args, " "))
	return exec.Command(args[0], args[1:]...)
}

// Git wraps shelling out to git, ensuring we're running from the git repository
//
// Fails immediately if any git command fails.
func (g *Repository) Git(command ...string) {
	g.T.Helper()

	cmd := g.gitCmd(command...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.Logger.Info(string(out))
		g.T.Fatal(err)
	}
}

// initialCommit initializes the Nomos repo with the Repo object.
func (g *Repository) initialCommit(sourceFormat filesystem.SourceFormat) {
	g.T.Helper()

	// Add .gitkeep to retain dir when empty, otherwise configsync will error.
	g.AddEmptyDir(AcmeDir)
	if g.Type == RootRepo {
		// Add safety Namespace and ClusterRole to avoid errors from the safety
		// check (KNV2006) when deleting all the other remaining objects.
		g.AddSafetyNamespace()
		g.AddSafetyClusterRole()
	}
	switch sourceFormat {
	case filesystem.SourceFormatHierarchy:
		// Hierarchy format requires a Repo object.
		g.Logger.Infof("[repo %s] Setting repo format to %s", path.Base(g.Root), sourceFormat)
		g.Add(filepath.Join(AcmeDir, "system", "repo.yaml"), fake.RepoObject())
	case filesystem.SourceFormatUnstructured:
		// It is an error for unstructured repos to include the Repo object.
		g.Logger.Infof("[repo %s] Setting repo format to %s", path.Base(g.Root), sourceFormat)
	default:
		g.T.Fatalf("Unrecognized SourceFormat: %q", sourceFormat)
	}
	g.Format = sourceFormat
	g.CommitAndPush("initial commit")
}

// init initializes this git repository and configures it to talk to the cluster
// under test.
func (g *Repository) init(privateKey string) {
	g.T.Helper()

	if err := os.RemoveAll(g.Root); err != nil {
		g.T.Fatal(err)
	}

	err := os.MkdirAll(g.Root, fileMode)
	if err != nil {
		g.T.Fatal(err)
	}
	g.Git("init")
	g.Git("checkout", "-b", MainBranch)

	// We have to configure username/email or else committing to the repository
	// produces errors.
	g.Git("config", "user.name", "E2E Testing")
	g.Git("config", "user.email", "team@example.com")

	// Use ssh rather than the default that git uses, as the default does not know
	// how to use private key files.
	g.Git("config", "ssh.variant", "ssh")
	// Overwrite the ssh command to:
	// 1) Not perform host key checking for git-server, since this isn't set up
	//   properly and we don't care.
	// 2) Use the private key file we generated.
	g.Git("config", "core.sshCommand",
		fmt.Sprintf("ssh -q -o StrictHostKeyChecking=no -i %s", privateKey))
}

// Add writes a YAML or JSON representation of obj to `path` in the git
// repository, and `git add`s the file. Does not commit/push.
//
// Overwrites the file if it already exists.
// Automatically writes YAML or JSON based on the path's extension.
//
// Don't put multiple manifests in the same file unless parsing multi-manifest
// files is the behavior under test. In that case, use AddFile.
func (g *Repository) Add(path string, obj client.Object) {
	g.T.Helper()
	testkubeclient.AddTestLabel(obj)
	// TODO: Figure out how to cleanly inject runtime.Scheme here.

	// We have to make a pass through json since yaml.Marshal does not respect
	// json "omitempty" directives.
	var bytes []byte
	var err error
	var u *unstructured.Unstructured
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		// We must convert through JSON/Unstructured to avoid "omitempty" fields
		// from being specified.
		u, err = reconcile.AsUnstructuredSanitized(obj)
		if err != nil {
			g.T.Fatal(err)
		}
		bytes, err = yaml.Marshal(u)
	case ".json":
		u, err = reconcile.AsUnstructuredSanitized(obj)
		if err != nil {
			g.T.Fatal(err)
		}
		bytes, err = json.MarshalIndent(u, "", "  ")
	default:
		// If you're seeing this error, use "AddFile" instead to test ignoring
		// files with extensions we ignore.
		err = fmt.Errorf("invalid extension to write object to, %q, use .AddFile() instead", ext)
	}
	if err != nil {
		g.T.Fatal(err)
	}

	g.AddFile(path, bytes)
}

// Get reads, parses, and returns the specified file as an object.
//
// File must have one of these suffixes: .yaml, .yml, .json
// This is meant to read files written with Add. So it only reads one object per
// file. If you need to parse multiple objects from one file, use GetFile.
func (g *Repository) Get(path string) client.Object {
	g.T.Helper()

	bytes := g.GetFile(path)

	var err error
	uObj := &unstructured.Unstructured{}
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml", ".json":
		err = yaml.Unmarshal(bytes, uObj)
	default:
		// If you're seeing this error, use "GetFile" instead to test ignoring
		// files with extensions we ignore.
		err = fmt.Errorf("invalid extension to read object from, %q, use .GetFile() instead", ext)
	}
	if err != nil {
		g.T.Fatal(err)
	}

	gvk := uObj.GroupVersionKind()
	if gvk.Empty() {
		g.T.Fatalf("missing GVK in file: %s", path)
	}

	// Lookup type by GVK of those registered with the Scheme
	tObj, err := g.Scheme.New(gvk)
	if err != nil {
		// Return the unstructured object if the GVK is not registered with the Scheme
		if runtime.IsNotRegisteredError(err) {
			return uObj
		}
		g.T.Fatal(err)
	}

	obj, ok := tObj.(client.Object)
	if !ok {
		// Return the unstructured object if the typed object is a
		// runtime.Object but not a metav1.Object.
		// Most registered objects should be metav1.Object.
		// But if they aren't, we need to be able to read their metadata.
		return uObj
	}

	// Convert the unstructured object to a typed object
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(uObj.UnstructuredContent(), obj)
	if err != nil {
		g.T.Fatal(err)
	}
	return obj
}

// GetAll reads, parses, and returns all the files in a specified directory as
// objects.
func (g *Repository) GetAll(dirPath string, recursive bool) []client.Object {
	g.T.Helper()

	absPath := filepath.Join(g.Root, dirPath)

	entries, err := os.ReadDir(absPath)
	if err != nil {
		g.T.Fatal(err)
	}

	var objs []client.Object
	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if recursive {
				objs = append(objs, g.GetAll(entryPath, recursive)...)
			}
		} else {
			ext := filepath.Ext(entry.Name())
			switch ext {
			case ".yaml", ".yml", ".json":
				objs = append(objs, g.Get(entryPath))
			default:
				// ignore files that aren't yaml or json
			}
		}
	}
	return objs
}

// AddFile writes `bytes` to `file` in the git repository.
// This function should only be directly used for testing the literal YAML/JSON
// parsing logic.
//
// Path is relative to the Git repository root.
// Overwrites `file` if it already exists.
// Does not commit/push.
func (g *Repository) AddFile(path string, bytes []byte) {
	g.T.Helper()

	absPath := filepath.Join(g.Root, path)

	err := os.MkdirAll(filepath.Dir(absPath), fileMode)
	if err != nil {
		g.T.Fatal(err)
	}

	// Write bytes to file.
	err = ioutil.WriteFile(absPath, bytes, fileMode)
	if err != nil {
		g.T.Fatal(err)
	}
	// Add the file to Git.
	g.Git("add", absPath)
}

// AddEmptyDir creates an empty dir containing an empty .gitkeep file, so the
// empty dir will be retained in git.
//
// Use this when creating empty sync directories, otherwise Config Sync will
// error that the directory doesn't exist.
func (g *Repository) AddEmptyDir(path string) {
	g.T.Helper()

	g.AddFile(filepath.Join(path, GitKeepFileName), []byte{})
}

// GetFile reads and returns the specified file.
func (g *Repository) GetFile(path string) []byte {
	g.T.Helper()

	absPath := filepath.Join(g.Root, path)

	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		g.T.Fatal(err)
	}

	return bytes
}

// Copy copies the file or directory from source to destination.
// Overwrites the file if it already exists.
// Does not commit/push.
func (g *Repository) Copy(sourceDir, destDir string) {
	g.T.Helper()

	absDestPath := filepath.Join(g.Root, destDir)
	parentDir := filepath.Dir(absDestPath)
	if absDestPath != parentDir {
		if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
			g.T.Fatalf("failed to create directory: %s", parentDir)
		}
	}
	if out, err := exec.Command("cp", "-r", sourceDir, absDestPath).CombinedOutput(); err != nil {
		g.T.Fatalf("failed to copy directory: %s", string(out))
	}
	// Add the directory to Git.
	g.Git("add", absDestPath)
}

// Remove deletes `file` from the git repository.
// If `file` is a directory, deletes the directory.
// Returns error if the file does not exist.
// Does not commit/push.
func (g *Repository) Remove(path string) {
	g.T.Helper()

	absPath := filepath.Join(g.Root, path)

	err := os.RemoveAll(absPath)
	if err != nil {
		g.T.Fatal(err)
	}

	g.Git("add", absPath)
}

// Exists returns true if the file or directory exists at the specified path.
func (g *Repository) Exists(path string) bool {
	g.T.Helper()

	absPath := filepath.Join(g.Root, path)

	_, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		g.T.Fatal(err)
	}
	return true
}

// CommitAndPush commits any changes to the git repository, and
// pushes them to the git server.
// We don't care about differentiating between committing and pushing
// for tests.
func (g *Repository) CommitAndPush(msg string) {
	g.T.Helper()
	g.CommitAndPushBranch(msg, MainBranch)
}

// CommitAndPushBranch commits any changes to the git branch, and
// pushes them to the git server.
func (g *Repository) CommitAndPushBranch(msg, branch string) {
	g.T.Helper()

	g.Git("commit", "-m", msg)

	g.Logger.Infof("[repo %s] committing %q (%s)", path.Base(g.Root), msg, g.Hash())
	g.Push(branch, "-f")
}

// Push pushes the provided refspec to the git server.
// Performs a retry using RemoteURL, which may change if the port forwarding restarts.
func (g *Repository) Push(refspec string, flags ...string) {
	var out []byte
	took, err := retry.Retry(1*time.Minute, func() error {
		args := []string{"push"}
		args = append(args, flags...)
		remoteURL, err := g.GitProvider.RemoteURL(g.RemoteRepoName)
		if err != nil {
			return err
		}
		args = append(args, remoteURL, refspec)
		cmd := g.gitCmd(args...)
		out, err = cmd.CombinedOutput()
		return err
	})
	if err != nil {
		g.T.Log(string(out))
		g.T.Fatalf("took %v, err %v", took, err)
	}
}

func (g *Repository) pushAllToRemote(remote string) {
	cmd := g.gitCmd("push", remote, "--all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.T.Log(string(out))
		g.T.Errorf("failed to push to remote %s: %v", remote, err)
	}
}

// CreateBranch creates and checkouts a new branch at once.
func (g *Repository) CreateBranch(branch string) {
	g.T.Helper()

	g.Git("branch", branch)
	g.CheckoutBranch(branch)
}

// CheckoutBranch checkouts a branch.
func (g *Repository) CheckoutBranch(branch string) {
	g.T.Helper()

	g.Git("checkout", branch)
}

// RenameBranch renames the current branch with a new one both locally and remotely.
// The old branch will be deleted from remote.
func (g *Repository) RenameBranch(current, new string) {
	g.T.Helper()

	g.Git("branch", "-m", current, new)
	g.Push(new)
	g.Push(current, "--delete")
}

// Hash returns the current hash of the git repository.
//
// Immediately ends the test on error.
func (g *Repository) Hash() string {
	// Get the hash of the git repository.
	// git rev-parse --verify HEAD
	out, err := g.gitCmd("rev-parse", "--verify", "HEAD").CombinedOutput()
	if err != nil {
		g.Logger.Info(string(out))
		g.T.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// AddSafetyNamespace adds a Namespace to prevent the mono-repo safety check
// (KNV2006) from preventing deletion of other objects.
func (g *Repository) AddSafetyNamespace() {
	g.T.Helper()
	g.Add(g.SafetyNSPath, fake.NamespaceObject(g.SafetyNSName))
}

// RemoveSafetyNamespace removes the safety Namespace.
func (g *Repository) RemoveSafetyNamespace() {
	g.T.Helper()
	g.Remove(g.SafetyNSPath)
}

// AddSafetyClusterRole adds a ClusterRole to prevent the mono-repo safety check
// (KNV2006) from preventing deletion of other objects.
func (g *Repository) AddSafetyClusterRole() {
	g.T.Helper()
	g.Add(g.SafetyClusterRolePath, fake.ClusterRoleObject(core.Name(g.SafetyClusterRoleName)))
}

// RemoveSafetyClusterRole removes the safety ClusterRole.
func (g *Repository) RemoveSafetyClusterRole() {
	g.T.Helper()
	g.Remove(g.SafetyClusterRolePath)
}
