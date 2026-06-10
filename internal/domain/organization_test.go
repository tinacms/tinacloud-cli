package domain

import "testing"

func TestOrganizationWithRoleString(t *testing.T) {
	tests := []struct {
		name string
		org  OrganizationWithRole
		want string
	}{
		{
			name: "role present",
			org: OrganizationWithRole{
				Organization: Organization{ID: "org-1", Name: "Acme", BillingEnabled: true},
				Role:         "admin",
			},
			want: "Acme (org-1) [Role: Admin] [Billing enabled]",
		},
		{
			name: "role empty",
			org: OrganizationWithRole{
				Organization: Organization{ID: "org-2", Name: "Widgets", BillingEnabled: false},
				Role:         "",
			},
			want: "Widgets (org-2) [Billing disabled]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.org.String()
			if got != tc.want {
				t.Errorf("OrganizationWithRole.String() = %q, want %q", got, tc.want)
			}
		})
	}
}
