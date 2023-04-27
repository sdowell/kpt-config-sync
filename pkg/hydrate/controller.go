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

package hydrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"kpt.dev/configsync/pkg/api/configsync"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	"kpt.dev/configsync/pkg/importer/filesystem/cmpath"
	"kpt.dev/configsync/pkg/metadata"
	"kpt.dev/configsync/pkg/reconcilermanager"
	"kpt.dev/configsync/pkg/status"
	"kpt.dev/configsync/pkg/util/log"
)

const (
	// tmpLink is the temporary soft link name.
	tmpLink = "tmp-link"
	// DoneFile is the file name that indicates the hydration is done.
	DoneFile = "done"
	// ErrorFile is the file name of the hydration errors.
	ErrorFile = "error.json"
)

// Hydrator runs the hydration process.
type Hydrator struct {
	// DonePath is the absolute path to the done file under the /repo directory.
	DonePath cmpath.Absolute
	// SourceType is the type of the source repository, must be git or oci.
	SourceType v1beta1.SourceType
	// SourceRoot is the absolute path to the source root directory.
	SourceRoot cmpath.Absolute
	// HydratedRoot is the absolute path to the hydrated root directory.
	HydratedRoot cmpath.Absolute
	// SourceLink is the name of (a symlink to) the source directory under SourceRoot, which contains the clone of the git repo.
	SourceLink string
	// HydratedLink is the name of (a symlink to) the source directory under HydratedRoot, which contains the hydrated configs.
	HydratedLink string
	// SyncDir is the relative path to the configs within the Git repository.
	SyncDir cmpath.Relative
	// PollingPeriod is the period of time between checking the filesystem for source updates to render.
	PollingPeriod time.Duration
	// RehydratePeriod is the period of time between rehydrating on errors.
	RehydratePeriod time.Duration
	// ReconcilerName is the name of the reconciler.
	ReconcilerName string
}

// Run runs the hydration process periodically.
func (h *Hydrator) Run(ctx context.Context) {
	// Use timers, not tickers.
	// Tickers can cause memory leaks and continuous execution, when execution
	// takes longer than the tick duration.
	runTimer := time.NewTimer(h.PollingPeriod)
	defer runTimer.Stop()
	rehydrateTimer := time.NewTimer(h.RehydratePeriod)
	defer rehydrateTimer.Stop()
	absSourceDir := h.absSourceDir()
	for {
		select {
		case <-ctx.Done():
			return
		case <-rehydrateTimer.C:
			commit, syncDir, err := SourceCommitAndDir(h.SourceType, absSourceDir, h.SyncDir, h.ReconcilerName)
			if err != nil {
				klog.Errorf("failed to get the commit hash and sync directory from the source directory %s: %v", absSourceDir.OSPath(), err)
			} else {
				h.rehydrateOnError(commit, syncDir.OSPath())
			}
			rehydrateTimer.Reset(h.RehydratePeriod) // Schedule rehydrate attempt
		case <-runTimer.C:
			commit, syncDir, err := SourceCommitAndDir(h.SourceType, absSourceDir, h.SyncDir, h.ReconcilerName)
			if err != nil {
				klog.Errorf("failed to get the commit hash and sync directory from the source directory %s: %v", absSourceDir.OSPath(), err)
			} else if DoneCommit(h.DonePath.OSPath()) != commit {
				// If the commit has been processed before, regardless of success or failure,
				// skip the hydration to avoid repeated execution.
				// The rehydrate ticker will retry on the failed commit.
				hydrateErr := h.hydrate(commit, syncDir.OSPath())
				if err := h.complete(commit, hydrateErr); err != nil {
					klog.Errorf("failed to complete the rendering execution for commit %q: %v", commit, err)
				}
			}
			runTimer.Reset(h.PollingPeriod) // Schedule re-run attempt
		}
	}
}

// runHydrate runs `kustomize build` on the source configs.
func (h *Hydrator) runHydrate(sourceCommit, syncDir string) HydrationError {
	newHydratedDir := h.HydratedRoot.Join(cmpath.RelativeOS(sourceCommit))
	dest := newHydratedDir.Join(h.SyncDir).OSPath()

	if err := kustomizeBuild(syncDir, dest, true); err != nil {
		return err
	}

	newCommit, err := ComputeCommit(h.absSourceDir())
	if err != nil {
		return NewTransientError(err)
	} else if sourceCommit != newCommit {
		return NewTransientError(fmt.Errorf("source commit changed while running Kustomize build, was %s, now %s. It will be retried in the next sync", sourceCommit, newCommit))
	}

	if err := updateSymlink(h.HydratedRoot.OSPath(), h.HydratedLink, newHydratedDir.OSPath()); err != nil {
		return NewInternalError(errors.Wrapf(err, "unable to update the symbolic link to %s", newHydratedDir.OSPath()))
	}
	klog.Infof("Successfully rendered %s for commit %s", syncDir, sourceCommit)
	return nil
}

// ComputeCommit returns the computed commit from given sourceDir, or error
// if the sourceDir fails symbolic link evaluation
func ComputeCommit(sourceDir cmpath.Absolute) (string, error) {
	dir, err := sourceDir.EvalSymlinks()
	if err != nil {
		return "", errors.Wrapf(err, "unable to evaluate the symbolic link of sourceDir %s", dir)
	}
	newCommit := filepath.Base(dir.OSPath())
	return newCommit, nil
}

// absSourceDir returns the absolute path of a source directory by joining the
// root source directory path and a relative path to the source directory
func (h *Hydrator) absSourceDir() cmpath.Absolute {
	return h.SourceRoot.Join(cmpath.RelativeSlash(h.SourceLink))
}

// hydrate renders the source git repo to hydrated configs.
func (h *Hydrator) hydrate(sourceCommit, syncDir string) HydrationError {
	hydrate, err := needsKustomize(syncDir)
	if err != nil {
		return NewInternalError(errors.Wrapf(err, "unable to check if rendering is needed for the source directory: %s", syncDir))
	}
	if !hydrate {
		found, err := hasKustomizeSubdir(syncDir)
		if err != nil {
			return NewInternalError(err)
		}
		if found {
			return NewActionableError(errors.Errorf("Kustomization config file is missing from the sync directory %s. "+
				"To fix, either add kustomization.yaml in the sync directory to trigger the rendering process, "+
				"or remove kustomizaiton.yaml from all sub directories to skip rendering.", syncDir))
		}
		klog.V(5).Infof("no rendering is needed because of no Kustomization config file in the source configs with commit %s", sourceCommit)
		if err := os.RemoveAll(h.HydratedRoot.OSPath()); err != nil {
			return NewInternalError(err)
		}
		return nil
	}

	// Remove the done file because a new hydration is in progress.
	if err := os.RemoveAll(h.DonePath.OSPath()); err != nil {
		return NewInternalError(errors.Wrapf(err, "unable to remove the done file: %s", h.DonePath.OSPath()))
	}
	return h.runHydrate(sourceCommit, syncDir)
}

// rehydrateOnError retries the hydration on errors.
func (h *Hydrator) rehydrateOnError(sourceCommit, syncDir string) {
	errorFile := h.HydratedRoot.Join(cmpath.RelativeSlash(ErrorFile))
	if _, err := os.Stat(errorFile.OSPath()); err != nil {
		if !os.IsNotExist(err) {
			klog.Warningf("unable to check the error file %s: %v", errorFile, err)
		}
		return
	}
	klog.Infof("retry rendering commit %s", sourceCommit)
	hydrationErr := h.runHydrate(sourceCommit, syncDir)
	if err := h.complete(sourceCommit, hydrationErr); err != nil {
		klog.Errorf("failed to complete the re-rendering execution for commit %q: %v", sourceCommit, err)
	}
}

// updateSymlink updates the symbolic link to the hydrated directory.
func updateSymlink(hydratedRoot, link, newDir string) error {
	linkPath := filepath.Join(hydratedRoot, link)
	tmpLinkPath := filepath.Join(hydratedRoot, tmpLink)

	oldDir, err := filepath.EvalSymlinks(linkPath)
	if oldDir == newDir {
		return nil
	}
	deleteOldDir := true
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "unable to access the current hydrated directory: %s", linkPath)
		}
		deleteOldDir = false
	}

	if err := os.Symlink(newDir, tmpLinkPath); err != nil {
		return errors.Wrap(err, "unable to create symlink")
	}

	if err := os.Rename(tmpLinkPath, linkPath); err != nil {
		return errors.Wrap(err, "unable to replace symlink")
	}

	if deleteOldDir {
		if err := os.RemoveAll(oldDir); err != nil {
			klog.Warningf("unable to remove the previously hydrated directory %s: %v", oldDir, err)
		}
	}
	return nil
}

// complete marks the hydration process is done with a done file under the /repo directory
// and reset the error file (create, update or delete).
func (h *Hydrator) complete(commit string, hydrationErr HydrationError) error {
	errorPath := h.HydratedRoot.Join(cmpath.RelativeSlash(ErrorFile)).OSPath()
	var err error
	if hydrationErr == nil {
		err = deleteErrorFile(errorPath)
	} else {
		err = exportError(commit, h.HydratedRoot.OSPath(), errorPath, hydrationErr)
	}
	if err != nil {
		return err
	}
	done, err := os.Create(h.DonePath.OSPath())
	if err != nil {
		return errors.Wrapf(err, "unable to create done file: %s", h.DonePath.OSPath())
	}
	if _, err = done.WriteString(commit); err != nil {
		return errors.Wrapf(err, "unable to write to commit hash to the done file: %s", h.DonePath)
	}
	if err := done.Close(); err != nil {
		klog.Warningf("unable to close the done file %s: %v", h.DonePath.OSPath(), err)
	}
	return nil
}

// DoneCommit extracts the commit hash from the done file if exists.
// It returns the commit hash if exists, otherwise, returns an empty string.
// If it fails to extract the commit hash for various errors, we only log a warning,
// and wait for the next hydration loop to retry the hydration.
func DoneCommit(donePath string) string {
	if _, err := os.Stat(donePath); err == nil {
		commit, err := ioutil.ReadFile(donePath)
		if err != nil {
			klog.Warningf("unable to read the done file %s: %v", donePath, err)
			return ""
		}
		return string(commit)
	} else if !os.IsNotExist(err) {
		klog.Warningf("unable to check the status of the done file %s: %v", donePath, err)
	}
	return ""
}

// exportError writes the error content to the error file.
func exportError(commit, root, errorFile string, hydrationError HydrationError) error {
	klog.Errorf("rendering error for commit %s: %v", commit, hydrationError)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fileMode := os.FileMode(0755)
		if err := os.Mkdir(root, fileMode); err != nil {
			return errors.Wrapf(err, "unable to create the root directory: %s", root)
		}
	}

	tmpFile, err := ioutil.TempFile(root, "tmp-err-")
	if err != nil {
		return errors.Wrapf(err, "unable to create temporary error-file under directory %s", root)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			klog.Warningf("unable to close temporary error-file: %s", tmpFile.Name())
		}
	}()

	payload := HydrationErrorPayload{
		Code:  hydrationError.Code(),
		Error: hydrationError.Error(),
	}

	jb, err := json.Marshal(payload)
	if err != nil {
		klog.Errorf("can't encode hydration error payload: %v", err)
		return err
	}

	if _, err = tmpFile.Write(jb); err != nil {
		return errors.Wrapf(err, "unable to write to temporary error-file: %s", tmpFile.Name())
	}
	if err := os.Rename(tmpFile.Name(), errorFile); err != nil {
		return errors.Wrapf(err, "unable to rename %s to %s", tmpFile.Name(), errorFile)
	}
	if err := os.Chmod(errorFile, 0644); err != nil {
		return errors.Wrapf(err, "unable to change permissions on the error-file: %s", errorFile)
	}
	klog.Infof("Saved the rendering error in file: %s", errorFile)
	return nil
}

// deleteErrorFile deletes the error file.
func deleteErrorFile(file string) error {
	if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "unable to delete error file: %s", file)
	}
	return nil
}

// filterArgsFromErrorPayload removes the Args parameter from ErrorPayload.
// The Args parameter sometimes contains fields which are not stable across
// iterations, such as failCount. This causes the error message to flap, which
// makes it impossible to use errors as a unique key for notification deliveries.
func filterArgsFromErrorPayload(content []byte) []byte {
	errorPayload := &log.ErrorPayload{}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(errorPayload); err != nil {
		klog.Warningf("error unmarshalling error payload: %v", err)
		return content
	}
	errorPayload.Args = nil
	newContent, err := json.Marshal(errorPayload)
	if err != nil {
		klog.Warningf("error marshalling error payload: %v", err)
		return content
	}
	return newContent
}

// SourceCommitAndDir returns the source hash (a git commit hash or an OCI image digest or a helm chart version), the absolute path of the sync directory, and source errors.
func SourceCommitAndDir(sourceType v1beta1.SourceType, sourceRevDir cmpath.Absolute, syncDir cmpath.Relative, reconcilerName string) (string, cmpath.Absolute, status.Error) {
	// Check if the source root directory is mounted
	sourceRoot := path.Dir(sourceRevDir.OSPath())
	if _, err := os.Stat(sourceRoot); err != nil && os.IsNotExist(err) {
		return "", "", status.TransientError(err)
	}
	// Check if the source configs are synced successfully.
	errFilePath := filepath.Join(sourceRoot, ErrorFile)

	var containerName string
	switch sourceType {
	case v1beta1.OciSource:
		containerName = reconcilermanager.OciSync
	case v1beta1.GitSource:
		containerName = reconcilermanager.GitSync
	case v1beta1.HelmSource:
		containerName = reconcilermanager.HelmSync
	}

	content, err := os.ReadFile(errFilePath)
	if err != nil && !os.IsNotExist(err) {
		return "", "", status.TransientError(
			fmt.Errorf("unable to load %s: %v. Please "+
				"check %s logs for more info: kubectl logs -n %s -l %s -c %s",
				errFilePath, err, containerName, configsync.ControllerNamespace,
				metadata.ReconcilerLabel, reconcilerName))
	} else if err == nil && len(content) == 0 {
		return "", "", status.SourceError.Sprintf("%s is "+
			"empty. Please check %s logs for more info: kubectl logs -n %s -l %s -c %s",
			errFilePath, containerName, configsync.ControllerNamespace,
			metadata.ReconcilerLabel, reconcilerName).Build()
	} else if err == nil {
		filteredContent := filterArgsFromErrorPayload(content)
		return "", "", status.SourceError.Sprintf("error in the %s container: %s", containerName, string(filteredContent)).Build()
	}

	gitDir, err := sourceRevDir.EvalSymlinks()
	if err != nil {
		return "", "", status.SourceError.Sprintf("unable to evaluate the source link %s", sourceRevDir).Wrap(err).Build()
	}

	commit := filepath.Base(gitDir.OSPath())

	// The hydration controller might pull remote Helm charts locally, which makes the source directory dirty.
	// Hence, we don't check if the source directory is clean before the hydration.
	// The assumption is that customers should have limited access to manually modify the source configs.
	// For the local Helm charts pulled by the Helm inflator, the entire hydrated directory
	// will blow away when new commits come in.
	// If the commit hash is not changed, the hydration will be skipped.
	// Therefore, it is relatively safe to keep the Helm charts local in the source directory.
	relSyncDir := gitDir.Join(syncDir)
	sourceDir, err := relSyncDir.EvalSymlinks()
	if err != nil {
		return commit, "", status.PathWrapError(
			errors.Wrap(err, "evaluating symbolic link to policy sourceRoot"), relSyncDir.OSPath())
	}
	return commit, sourceDir, nil
}
