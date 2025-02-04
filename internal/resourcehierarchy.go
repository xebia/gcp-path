package internal

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/binxio/gcloudconfig"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	asset "cloud.google.com/go/asset/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
)

type ResourceHierarchy struct {
	Organizations []*OrganizationNode
}

func LoadResourceHierarchy(ctx context.Context, displayNames []string, viaResourceManager bool) (*ResourceHierarchy, error) {
	var err error

	var credentials *google.Credentials
	if gcloudconfig.IsGCloudOnPath() {
		if credentials, err = gcloudconfig.GetCredentials(""); err != nil {
			return nil, err
		}
	} else {
		if credentials, err = google.FindDefaultCredentials(ctx); err == nil {
			return nil, err
		}
	}
	organizationsClient, err := resourcemanager.NewOrganizationsClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, err
	}

	foldersClient, err := resourcemanager.NewFoldersClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, err
	}

	assetClient, err := asset.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, err
	}

	organizations, err := ListOrganizations(ctx, organizationsClient, displayNames)
	if err != nil {
		return nil, err
	}

	for _, organization := range organizations {
		if viaResourceManager {
			if err := organization.LoadFolderViaResourceManager(ctx, foldersClient, nil); err != nil {
				return nil, err
			}
		} else {
			if err := organization.LoadFoldersViaCloudAsset(ctx, assetClient); err != nil {
				return nil, err
			}
		}
	}
	return &ResourceHierarchy{Organizations: organizations}, nil
}

func (h *ResourceHierarchy) AvailableOrganizations() []string {
	result := make([]string, len(h.Organizations))
	for i, organization := range h.Organizations {
		result[i] = organization.Organization.DisplayName
	}
	return result
}

func (h *ResourceHierarchy) GetResourceName(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	if u.Scheme != "" {
		return "", errors.New("unsupported scheme")
	}
	if u.Host == "" {
		return "", fmt.Errorf("organization name is missing from path '%s'", path)
	}

	var organization *OrganizationNode
	for _, o := range h.Organizations {
		if o.Organization.DisplayName == u.Host {
			organization = o
			break
		}
	}
	if organization == nil {
		return "", fmt.Errorf("Organization '%s' does not exist or you have no access", u.Host)
	}
	return organization.GetResourceName(u.Path)
}

func (h *ResourceHierarchy) GetPathByResourceName(resourceName string) (string, error) {
	if strings.HasPrefix(resourceName, "organizations/") {
		for _, organization := range h.Organizations {
			if organization.Organization.Name == resourceName {
				return "//" + PathEscape(organization.Organization.DisplayName), nil
			}
			return "", fmt.Errorf("organization with resource name '%s' not found", resourceName)
		}
	}

	if strings.HasPrefix(resourceName, "folders/") {
		for _, organization := range h.Organizations {
			folder, ok := organization.Folders[resourceName]
			if ok {
				return folder.Path(), nil
			}
			return "", fmt.Errorf("folder with resource name '%s' not found", resourceName)
		}
	}
	return "", fmt.Errorf("unsupported resource name '%s'", resourceName)
}

func (h *ResourceHierarchy) GetOrganizationByName(name string) (*OrganizationNode, error) {
	for _, organization := range h.Organizations {
		if organization.Organization.DisplayName == name {
			return organization, nil
		}
	}
	return nil, fmt.Errorf("organization '%s' not found or you have no access", name)
}

func (h *ResourceHierarchy) ListPaths(names []string) error {
	for _, name := range names {
		organization, err := h.GetOrganizationByName(name)
		if err != nil {
			return err
		}
		for _, path := range organization.Paths() {
			fmt.Printf("%s\n", path)
		}
	}
	return nil
}
