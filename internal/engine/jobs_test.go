// Copyright 2023 Adevinta

package engine

import (
	"fmt"
	"testing"

	"github.com/adevinta/vulcan-agent/jobrunner"
	types "github.com/adevinta/vulcan-types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/lava/internal/checktype"
	"github.com/adevinta/lava/internal/config"
)

func TestGenerateChecks(t *testing.T) {
	tests := []struct {
		name       string
		checktypes checktype.Catalog
		targets    []config.Target
		want       []check
		wantNilErr bool
	}{
		{
			name: "one checktype and one target",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
			},
			want: []check{
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"DomainName",
						},
					},
					target: config.Target{
						Identifier: "example.com",
						AssetType:  types.DomainName,
					},
					options: map[string]any{},
				},
			},
			wantNilErr: true,
		},
		{
			name: "target overrides checktype options",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
					Options: map[string]interface{}{
						"option1": "checktype value 1",
						"option2": "checktype value 2",
						"option3": "checktype value 3",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
					Options: map[string]interface{}{
						"option2": "target value 2",
					},
				},
			},
			want: []check{
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"DomainName",
						},
						Options: map[string]interface{}{
							"option1": "checktype value 1",
							"option2": "checktype value 2",
							"option3": "checktype value 3",
						},
					},
					target: config.Target{
						Identifier: "example.com",
						AssetType:  types.DomainName,
						Options: map[string]interface{}{
							"option2": "target value 2",
						},
					},
					options: map[string]interface{}{
						"option1": "checktype value 1",
						"option2": "target value 2",
						"option3": "checktype value 3",
					},
				},
			},
			wantNilErr: true,
		},
		{
			name: "two checktypes and one target",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
				"checktype2": {
					Name:        "checktype2",
					Description: "checktype2 description",
					Image:       "namespace2/repository2:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
			},
			want: []check{
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"DomainName",
						},
					},
					target: config.Target{
						Identifier: "example.com",
						AssetType:  types.DomainName,
					},
					options: map[string]any{},
				},
				{
					checktype: checktype.Checktype{
						Name:        "checktype2",
						Description: "checktype2 description",
						Image:       "namespace2/repository2:tag",
						Assets: []string{
							"DomainName",
						},
					},
					target: config.Target{
						Identifier: "example.com",
						AssetType:  types.DomainName,
					},
					options: map[string]any{},
				},
			},
			wantNilErr: true,
		},
		{
			name: "incompatible target",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.GitRepository,
				},
			},
			want:       nil,
			wantNilErr: true,
		},
		{
			name: "invalid target asset type",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"Hostname",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  "InvalidAssetType",
				},
			},
			want:       nil,
			wantNilErr: false,
		},
		{
			name:       "no checktypes",
			checktypes: nil,
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.GitRepository,
				},
			},
			want:       nil,
			wantNilErr: true,
		},
		{
			name: "no targets",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets:    nil,
			want:       nil,
			wantNilErr: true,
		},
		{
			name: "target without asset type",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
				},
			},
			want:       nil,
			wantNilErr: false,
		},
		{
			name: "one checktype with two asset types and one target",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"Hostname",
						"WebAddress",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "www.example.com",
					AssetType:  types.Hostname,
				},
			},
			want: []check{
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"Hostname",
							"WebAddress",
						},
					},
					target: config.Target{
						Identifier: "www.example.com",
						AssetType:  types.Hostname,
					},
					options: map[string]any{},
				},
			},
			wantNilErr: true,
		},
		{
			name: "one checktype with two asset types and one target identifier with two asset types",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"Hostname",
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
				{
					Identifier: "example.com",
					AssetType:  types.Hostname,
				},
			},
			want: []check{
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"Hostname",
							"DomainName",
						},
					},
					target: config.Target{
						Identifier: "example.com",
						AssetType:  types.Hostname,
					},
					options: map[string]any{},
				},
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"Hostname",
							"DomainName",
						},
					},
					target: config.Target{
						Identifier: "example.com",
						AssetType:  types.DomainName,
					},
					options: map[string]any{},
				},
			},
			wantNilErr: true,
		},
		{
			name: "one target identifier with two asset types",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"Hostname",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "https://www.example.com",
					AssetType:  types.Hostname,
				},
				{
					Identifier: "https://www.example.com",
					AssetType:  types.WebAddress,
				},
			},
			want: []check{
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"Hostname",
						},
					},
					target: config.Target{
						Identifier: "https://www.example.com",
						AssetType:  types.Hostname,
					},
					options: map[string]any{},
				},
			},
			wantNilErr: true,
		},
		{
			name: "duplicated targets",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
			},
			want: []check{
				{
					checktype: checktype.Checktype{
						Name:        "checktype1",
						Description: "checktype1 description",
						Image:       "namespace/repository:tag",
						Assets: []string{
							"DomainName",
						},
					},
					target: config.Target{
						Identifier: "example.com",
						AssetType:  types.DomainName,
					},
					options: map[string]any{},
				},
			},
			wantNilErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateChecks(tt.checktypes, tt.targets)
			if (err == nil) != tt.wantNilErr {
				t.Fatalf("unexpected error value: %v", err)
			}
			diffOpts := []cmp.Option{
				cmp.AllowUnexported(check{}),
				cmpopts.SortSlices(checkLess),
				cmpopts.IgnoreFields(check{}, "id"),
			}
			if diff := cmp.Diff(tt.want, got, diffOpts...); diff != "" {
				t.Errorf("checks mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestGenerateJobs(t *testing.T) {
	tests := []struct {
		name       string
		checktypes checktype.Catalog
		targets    []config.Target
		want       []jobrunner.Job
		wantNilErr bool
	}{
		{
			name: "one checktype and one target",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
			},
			want: []jobrunner.Job{
				{
					Image:     "namespace/repository:tag",
					Target:    "example.com",
					AssetType: "DomainName",
					Options:   "{}",
				},
			},
			wantNilErr: true,
		},
		{
			name: "two checktypes and one target",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
				},
				"checktype2": {
					Name:        "checktype2",
					Description: "checktype2 description",
					Image:       "namespace2/repository2:tag",
					Assets: []string{
						"DomainName",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
			},
			want: []jobrunner.Job{
				{
					Image:     "namespace/repository:tag",
					Target:    "example.com",
					AssetType: "DomainName",
					Options:   "{}",
				},
				{
					Image:     "namespace2/repository2:tag",
					Target:    "example.com",
					AssetType: "DomainName",
					Options:   "{}",
				},
			},
			wantNilErr: true,
		},
		{
			name: "one checktype and one target with valid required vars",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
					RequiredVars: []any{
						"REQUIRED_VAR_1",
						"REQUIRED_VAR_2",
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
			},
			want: []jobrunner.Job{
				{
					Image:     "namespace/repository:tag",
					Target:    "example.com",
					AssetType: "DomainName",
					Options:   "{}",
					RequiredVars: []string{
						"REQUIRED_VAR_1",
						"REQUIRED_VAR_2",
					},
				},
			},
			wantNilErr: true,
		},
		{
			name: "one checktype and one target with invalid required vars",
			checktypes: checktype.Catalog{
				"checktype1": {
					Name:        "checktype1",
					Description: "checktype1 description",
					Image:       "namespace/repository:tag",
					Assets: []string{
						"DomainName",
					},
					RequiredVars: []int{
						1,
						2,
					},
				},
			},
			targets: []config.Target{
				{
					Identifier: "example.com",
					AssetType:  types.DomainName,
				},
			},
			want:       nil,
			wantNilErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateJobs(tt.checktypes, tt.targets)
			if (err == nil) != tt.wantNilErr {
				t.Fatalf("unexpected error value: %v", err)
			}
			diffOpts := []cmp.Option{
				cmpopts.SortSlices(jobLess),
				cmpopts.IgnoreFields(jobrunner.Job{}, "CheckID"),
			}
			if diff := cmp.Diff(tt.want, got, diffOpts...); diff != "" {
				t.Errorf("checks mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func checkLess(a, b check) bool {
	h := func(c check) string {
		c.id = ""
		return fmt.Sprintf("%#v", c)
	}
	return h(a) < h(b)
}

func jobLess(a, b jobrunner.Job) bool {
	h := func(j jobrunner.Job) string {
		j.CheckID = ""
		return fmt.Sprintf("%#v", j)
	}
	return h(a) < h(b)
}
