package internal

import (
	"log"
)

type Folder struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"displayName"`
	Ancestors    []string `json:"ancestors"`
	organization *OrganizationNode
}

func (f *Folder) IsPathMatch(path []string) bool {
	if len(path)+1 != len(f.Ancestors) {
		return false
	}
	for i := range path {
		resourceName := f.Ancestors[len(path)-i-1]
		p := f.organization.Folders[resourceName]
		if p.DisplayName != path[i] {
			return false
		}
	}
	return true
}

func (f *Folder) Path() string {
	result := "//" + PathEscape(f.organization.Organization.DisplayName)
	for i := len(f.Ancestors) - 2; i >= 0; i-- {
		parent, ok := f.organization.Folders[f.Ancestors[i]]
		if !ok {
			log.Fatalf("Ancestor '%s' not found", f.Ancestors[i])
		}
		result = result + "/" + PathEscape(parent.DisplayName)

	}
	return result
}
