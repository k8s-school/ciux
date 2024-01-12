package internal

import "fmt"

type GitRevision struct {
	Tag     string
	Counter int
	Hash    string
	Dirty   bool
	Branch  string
}

// GetVersion returns the reference as 'git describe ' will do
// except that tag is the latest semver annotated tag
func (rev *GitRevision) GetVersion() string {
	var dirty string
	if rev.Dirty {
		dirty = "-dirty"
	}
	var counterHash string
	if rev.Counter != 0 {
		counterHash = fmt.Sprintf("-%d-g%s", rev.Counter, rev.Hash[0:7])
	}
	tag := rev.Tag
	if rev.Tag == "" {
		tag = "v0"
	}
	version := fmt.Sprintf("%s%s%s", tag, counterHash, dirty)
	return version
}

func (rev *GitRevision) UpgradeTag() (string, error) {
	// Get the latest tag
	tag := rev.Tag
	if tag == "" {
		return "v0", nil
	}
	// Upgrade the tag
	semver := SemVerParse(tag)
	if semver == nil {
		return "", fmt.Errorf("invalid semver tag %s", tag)
	}
	rcId, err := semver.ParseReleaseCandidate()
	if err != nil {
		return "", fmt.Errorf("unable to parse release candidate: %v", err)
	}
	if rcId == -1 {
		semver.Patch++
		semver.Prerelease = []string{"rc0"}
		return semver.String(), nil
	} else {
		semver.Prerelease[0] = fmt.Sprintf("rc%d", rcId+1)
	}
	return semver.String(), nil
}
