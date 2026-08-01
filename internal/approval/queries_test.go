package approval

import "testing"

func TestMatchedPolicySummary(t *testing.T) {
	tests := []struct {
		name     string
		snapshot []byte
		want     *MatchedPolicySummary
	}{
		{
			name: "matched policy",
			snapshot: []byte(`{
				"policy_id": "3f2a58fd-5341-42e7-8b73-c734e703fb81",
				"policy_version_id": "19f42a4d-f72f-4448-a958-d27491b07cff",
				"name": "Refund approval",
				"priority": 100,
				"version_number": 2,
				"effect": "require_approval",
				"conditions": [{"field":"action.type","operator":"equals","value":"customer.refund"}],
				"approval_settings": {
					"deadline_seconds": 300,
					"approver_group_id": "06f3a83d-ea3c-4af6-bb78-6939b76b69ba"
				}
			}`),
			want: &MatchedPolicySummary{
				ID:              "3f2a58fd-5341-42e7-8b73-c734e703fb81",
				VersionID:       "19f42a4d-f72f-4448-a958-d27491b07cff",
				Name:            "Refund approval",
				Effect:          "require_approval",
				Priority:        100,
				VersionNumber:   2,
				DeadlineSeconds: 300,
			},
		},
		{
			name: "workspace default has no matched policy",
			snapshot: []byte(`{
				"name": "Workspace default",
				"effect": "require_approval"
			}`),
			want: nil,
		},
		{
			name:     "malformed snapshot",
			snapshot: []byte(`{"policy_id":`),
			want:     nil,
		},
		{
			name:     "empty snapshot",
			snapshot: nil,
			want:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := matchedPolicySummary(test.snapshot)
			if test.want == nil {
				if got != nil {
					t.Fatalf("expected nil summary, got %+v", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected matched policy summary, got nil")
			}
			if *got != *test.want {
				t.Fatalf("summary mismatch\n got: %+v\nwant: %+v", *got, *test.want)
			}
		})
	}
}
