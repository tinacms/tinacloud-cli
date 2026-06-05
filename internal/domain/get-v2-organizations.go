package domain

import (
	"fmt"
	"strings"
)

type OrganizationWithRole struct {
	Organization Organization `json:"organization"`
	Role         string       `json:"role"`
}

func (o OrganizationWithRole) String() string {
	var output strings.Builder

	fmt.Fprintf(&output, "%s (%s) [Role: %s]", o.Organization.Name, o.Organization.ID, strings.ToTitle(o.Role[0:1])+o.Role[1:])
	if o.Organization.BillingEnabled {
		fmt.Fprint(&output, " [Billing enabled]")
	} else {
		fmt.Fprint(&output, " [Billing disabled]")
	}

	return output.String()
}
