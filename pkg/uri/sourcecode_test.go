// Author: lipixun
// Created Time : 日 10/23 18:31:12 2016
//
// File Name: sourcecode_test.go
// Description:
//
package uri

import (
	"testing"
)

var (
	repositoryUriCases = []struct {
		Source    string
		Stringify string
		Good      bool
		Uri       RepositoryUri
	}{
		{
			Source: "",
			Good:   true,
		},
		{
			Source:    "repouri",
			Stringify: "repouri",
			Good:      true,
			Uri:       RepositoryUri{Uri: "repouri"},
		},
		{
			Source: "repouri///",
			Good:   false,
		},
		{
			Source:    "repouri///@branch",
			Stringify: "repouri///@branch",
			Good:      true,
			Uri:       RepositoryUri{Uri: "repouri", Branch: "branch"},
		},
		{
			Source:    "repouri///=commit",
			Stringify: "repouri///=commit",
			Good:      true,
			Uri:       RepositoryUri{Uri: "repouri", Commit: "commit"},
		},
	}

	targetUriCases = []struct {
		Source    string
		Stringify string
		Good      bool
		Uri       TargetUri
	}{
		{
			Source: "",
			Good:   true,
		},
		{
			Source:    "target",
			Stringify: "target",
			Good:      true,
			Uri:       TargetUri{Name: "target"},
		},
		{
			Source:    "::target",
			Stringify: "target",
			Good:      true,
			Uri:       TargetUri{Name: "target"},
		},
		{
			Source:    "repouri::target",
			Stringify: "repouri::target",
			Good:      true,
			Uri:       TargetUri{Repository: &RepositoryUri{Uri: "repouri"}, Name: "target"},
		},
		{
			Source: "repouri///::target",
			Good:   false,
		},
		{
			Source:    "repouri///@branch",
			Stringify: "repouri///@branch",
			Good:      true,
			Uri:       TargetUri{Repository: &RepositoryUri{Uri: "repouri", Branch: "branch"}},
		},
		{
			Source:    "repouri///=commit",
			Stringify: "repouri///=commit",
			Good:      true,
			Uri:       TargetUri{Repository: &RepositoryUri{Uri: "repouri", Commit: "commit"}},
		},
	}
)

func TestRepositoryUri(t *testing.T) {
	for _, tCase := range repositoryUriCases {
		r := ParseRepositoryUri(tCase.Source)
		if tCase.Good {
			if r == nil {
				t.Errorf("Incorrect result. Expect [%s], failed to parse", tCase.Uri.String())
				continue
			}
			// Test equal
			if !r.Equal(&tCase.Uri) {
				t.Errorf("Incorrect result. Expect [%s] Actual [%s]", tCase.Uri.String(), r.String())
				continue
			}
			// Test stringify
			if r.String() != tCase.Stringify {
				t.Errorf("Incorrect stringify result. Expect [%s] Actual [%s]", tCase.Stringify, r.String())
				continue
			}
		} else {
			if r != nil {
				t.Errorf("Uri [%s] should be a bad uri", tCase.Source)
				continue
			}
		}
	}
}

func TestTargetUri(t *testing.T) {
	for _, tCase := range targetUriCases {
		r := ParseTargetUri(tCase.Source)
		if tCase.Good {
			if r == nil {
				t.Errorf("Incorrect result. Expect [%s], failed to parse", tCase.Uri.String())
				continue
			}
			// Test equal
			if !r.Equal(&tCase.Uri) {
				t.Errorf("Incorrect result. Expect [%s] Actual [%s]", tCase.Uri.String(), r.String())
				continue
			}
			// Test stringify
			if r.String() != tCase.Stringify {
				t.Errorf("Incorrect stringify result. Expect [%s] Actual [%s]", tCase.Stringify, r.String())
				continue
			}
		} else {
			if r != nil {
				t.Errorf("Uri [%s] should be a bad uri", tCase.Source)
				continue
			}
		}
	}
}
