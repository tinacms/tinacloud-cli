package domain

import "fmt"

type Branch string

type TinaVersion struct {
	Major string `json:"major"`
	Minor string `json:"minor"`
	Patch string `json:"patch"`
}

func (tv TinaVersion) String() string {
	return fmt.Sprintf("%s.%s.%s", tv.Major, tv.Minor, tv.Patch)
}

type ProjectMetadata struct {
	TinaVersion TinaVersion `json:"tinaVersion"`
}
