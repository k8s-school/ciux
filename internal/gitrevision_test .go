package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name     string
		rev      GitRevision
		expected string
	}{
		{
			name: "simple",
			rev: GitRevision{
				Tag:      "v1.0.0",
				Counter:  1,
				HeadHash: "1234567890abcdef",
				Dirty:    false,
			},
			expected: "v1.0.0-1-g1234567",
		},
		{
			name: "dirty",
			rev: GitRevision{
				Tag:      "v1.0.0",
				Counter:  1,
				HeadHash: "1234567890abcdef",
				Dirty:    true,
			},
			expected: "v1.0.0-1-g1234567-dirty",
		},
		{
			name: "tag",
			rev: GitRevision{
				Tag:      "v1.0.0",
				Counter:  0,
				HeadHash: "1234567890abcdef",
				Dirty:    false,
			},
			expected: "v1.0.0",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			actual := tt.rev.GetVersion()
			require.Equal(tt.expected, actual)
		})
	}

}

func TestUpgradeTag(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name           string
		revision       GitRevision
		expectedResult string
		expectedError  error
	}{
		{
			name: "no_tag",
			revision: GitRevision{
				Tag:      "",
				Counter:  0,
				HeadHash: "1234567890abcdef",
				Dirty:    false,
			},
			expectedResult: "v0.0.1-rc0",
			expectedError:  nil,
		},
		{
			name: "upgrade_patch",
			revision: GitRevision{
				Tag:      "v1.0.0",
				Counter:  0,
				HeadHash: "1234567890abcdef",
				Dirty:    false,
			},
			expectedResult: "v1.0.1-rc0",
			expectedError:  nil,
		},
		{
			name: "upgrade_rc",
			revision: GitRevision{
				Tag:      "v1.0.0-rc0",
				Counter:  0,
				HeadHash: "1234567890abcdef",
				Dirty:    false,
			},
			expectedResult: "v1.0.0-rc1",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result, err := tt.revision.UpgradeTag()

			require.Equal(tt.expectedResult, result)
			require.Equal(tt.expectedError, err)
		})
	}
}
