package internal

import (
	"fmt"
	"log/slog"
	"slices"

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
		slog.Debug("File patch : ", "from", from, "to", to)
		if from != nil {
			slog.Info("File patch", "from", from.Path())
			codeChange, err := IsFileInSourcePathes(from.Path(), pathes)
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
			codeChange, err := IsFileInSourcePathes(to.Path(), pathes)
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
//   - the latest ancestor commit for which the source code has changed, then list of following commits for which the source code has not changed
//   - an error if any
func FindCodeChange(repository *git.Repository, fromHash plumbing.Hash, pathes []string) ([]plumbing.Hash, error) {
	current, err := repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve commit from hash: %v", err)
	}
	hashes := []plumbing.Hash{}

	// Only one commit in the repository
	if len(current.ParentHashes) == 0 {
		hashes = append(hashes, current.Hash)
		return hashes, nil
	}

	var parent *object.Commit
	for len(current.ParentHashes) != 0 {
		parent, err = current.Parent(0)
		if err != nil {
			return hashes, fmt.Errorf("unable to retrieve parent commit: %v", err)
		}
		slog.Info("Current commit", "hash", current.Hash)
		slog.Info("Parent commit", "hash", parent.Hash)
		changed, err := HasDiff(current, parent, pathes)
		if err != nil {
			return hashes, err
		} else if changed {
			hashes = append(hashes, current.Hash)
			break
		}
		hashes = append(hashes, current.Hash)
		current = parent
	}
	slices.Reverse(hashes)
	return hashes, nil
}
