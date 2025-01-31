package internal

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"

	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/structpb"
)

type OrganizationNode struct {
	Organization *resourcemanagerpb.Organization
	Folders      map[string]*Folder
}

func (o *OrganizationNode) Paths() []string {
	result := []string{}
	for _, folder := range o.Folders {
		result = append(result, folder.Path())
	}
	return result
}

func ListOrganizations(ctx context.Context, client *resourcemanager.OrganizationsClient, displayNames []string) ([]*OrganizationNode, error) {
	var err error
	var organization *resourcemanagerpb.Organization
	organizations := make([]*OrganizationNode, 0)

	request := &resourcemanagerpb.SearchOrganizationsRequest{}
	it := client.SearchOrganizations(ctx, request)
	for organization, err = it.Next(); err == nil; organization, err = it.Next() {
		if displayNames == nil || len(displayNames) == 0 || slices.Contains(displayNames, organization.DisplayName) {
		}
		organizations = append(organizations, &OrganizationNode{Organization: organization, Folders: nil})
	}
	if errors.Is(err, iterator.Done) {
		return organizations, nil
	} else {
		return nil, err
	}
}

func (o *OrganizationNode) MarshalFolderFromStruct(s *structpb.Struct) (*Folder, error) {
	var result Folder
	if fValue, ok := s.Fields["f"]; ok {
		if listValue := fValue.GetListValue(); listValue != nil {
			if len(listValue.Values) != 3 {
				return nil, errors.New("expected 3 values in the row")
			}
			if nameValue, ok := listValue.Values[0].GetStructValue().Fields["v"]; ok {
				// google assets names, are namespaced but the ancestors values are not.
				result.Name, ok = strings.CutPrefix(nameValue.GetStringValue(), "//cloudresourcemanager.googleapis.com/")
				if !ok {
					return nil, fmt.Errorf("expected name '%s' to be prefixed with //cloudresourcemanager.googleapis.com/", nameValue.GetStringValue())
				}
			}
			if displayNameValue, ok := listValue.Values[1].GetStructValue().Fields["v"]; ok {
				result.DisplayName = displayNameValue.GetStringValue()
			} else {
				return nil, errors.New("missing displayName field")
			}
			if ancestorsValue, ok := listValue.Values[2].GetStructValue().Fields["v"]; ok {
				if ancestorsList := ancestorsValue.GetListValue(); ancestorsList != nil {
					for _, item := range ancestorsList.Values {
						result.Ancestors = append(result.Ancestors, item.GetStructValue().Fields["v"].GetStringValue())
					}
				}
			} else {
				return nil, errors.New("missing ancestors field")
			}
			result.organization = o
		}
	}
	return &result, nil
}

func (o *OrganizationNode) LoadFolders(ctx context.Context, client *asset.Client) error {
	var err error
	var response *assetpb.QueryAssetsResponse
	response, err = client.QueryAssets(ctx, &assetpb.QueryAssetsRequest{
		Parent: o.Organization.Name,
		Query: &assetpb.QueryAssetsRequest_Statement{
			Statement: "SELECT name, resource.data.displayName, ancestors FROM `cloudresourcemanager_googleapis_com_Folder`",
		},
	})
	if err != nil {
		return err
	}
	o.Folders = make(map[string]*Folder, response.GetQueryResult().TotalRows)
	for _, r := range response.GetQueryResult().GetRows() {
		folder, err := o.MarshalFolderFromStruct(r)
		if err != nil {
			return err
		}
		o.Folders[folder.Name] = folder
	}

	for !response.Done {
		response, err = client.QueryAssets(ctx, &assetpb.QueryAssetsRequest{
			Parent:    o.Organization.Name,
			PageToken: response.GetQueryResult().NextPageToken,
			Query:     &assetpb.QueryAssetsRequest_JobReference{response.JobReference},
		})
		if err != nil {
			return err
		}
		for _, r := range response.GetQueryResult().GetRows() {
			folder, err := o.MarshalFolderFromStruct(r)
			if err != nil {
				return err
			}
			o.Folders[folder.Name] = folder
		}
	}
	return nil
}

func (o *OrganizationNode) GetResourceName(path string) (string, error) {
	result := make([]*Folder, 0, 2)

	if path == "/" || path == "" {
		return o.Organization.Name, nil
	}
	pathParts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for _, folder := range o.Folders {
		if folder.IsPathMatch(pathParts) {
			result = append(result, folder)
		}
	}
	switch len(result) {
	case 0:
		return "", fmt.Errorf("no folder found with path '%s' in '%s'", path, o.Organization.DisplayName)
	case 1:

		return result[0].Name, nil
	default:
		return "", fmt.Errorf("multiple folders found with path '%s' in '%s'", path, o.Organization.DisplayName)
	}
}
