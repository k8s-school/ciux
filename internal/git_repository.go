package internal

import (
	"fmt"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func HasDiff(current *object.Commit, ancestor *object.Commit, pathes []string) (bool, error) {
	patch, err := ancestor.Patch(current)
	if err != nil {
		return false, fmt.Errorf("unable to get patch: %v", err)
	}
	// Get the file patch
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		if from != nil {
			slog.Info("File patch", "from", from.Path())
			codeChange, err := IsPathInSubdirectories(from.Path(), pathes)
			if err != nil {
				return false, err
			}
			if codeChange {
				if to != nil {
					slog.Debug("Source file changed", "path", to.Path())
				} else {
					slog.Debug("Source file removed", "path", from.Path())
				}
				return true, nil
			}
		} else if to != nil {
			slog.Info("File patch", "to", to.Path())
			codeChange, err := IsPathInSubdirectories(to.Path(), pathes)
			if err != nil {
				return false, err
			}
			if codeChange {
				slog.Debug("Source file changed", "path", to.Path())
				return true, nil
			}
		}
	}
	return false, nil
}

// FindCodeChange returns:
// - the latest ancestor commit where the source code has changed, or the first commit of the repository
// - the list of commits where the source code has not changed
// - an error if any
func FindCodeChange(repository *git.Repository, fromHash plumbing.Hash, pathes []string) (plumbing.Hash, []plumbing.Hash, error) {
	commit, _ := repository.CommitObject(fromHash)
	// First commit
	if len(commit.ParentHashes) == 0 {
		return fromHash, []plumbing.Hash{}, nil
	}

	var parent *object.Commit
	untouchedCodeCommit := make([]plumbing.Hash, 0)
	current := commit
	var err error
	for {
		parent, err = current.Parent(0)
		if err != nil {
			return plumbing.Hash{}, untouchedCodeCommit, fmt.Errorf("unable to retrieve parent commit: %v", err)
		}
		slog.Info("Ancestor commit", "hash", parent.Hash)
		changed, err := HasDiff(current, parent, pathes)
		if err != nil {
			return plumbing.Hash{}, untouchedCodeCommit, nil
		} else if changed {
			return current.Hash, untouchedCodeCommit, nil
		} else if len(parent.ParentHashes) == 0 {
			return parent.Hash, untouchedCodeCommit, nil
		}
		untouchedCodeCommit = append(untouchedCodeCommit, current.Hash)
		current = parent
	}
}
