package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xebia/gcp-path/internal"
)

func main() {
	root := &cobra.Command{Use: "gcp-path"}
	ls := &cobra.Command{
		Use:          "ls [organization name]....",
		Short:        "List resources",
		Args:         cobra.MinimumNArgs(0),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			hierarchy, err := internal.LoadResourceHierarchy(ctx, args)
			if err != nil {
				return err
			}
			return hierarchy.ListPaths(hierarchy.AvailableOrganizations())
		},
	}
	getResourceName := &cobra.Command{
		Use:          "get-resource-name [path...]",
		Short:        "Get resource name by path",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			hierarchy, err := internal.LoadResourceHierarchy(ctx, nil)
			if err != nil {
				return err
			}
			for _, path := range args {
				resourceName, err := hierarchy.GetResourceName(path)
				if err != nil {
					return err
				}
				idOnly, err := cmd.Flags().GetBool("id")
				if err != nil {
					return err
				}
				if idOnly {
					parts := strings.Split(resourceName, "/")
					resourceName = parts[len(parts)-1]
				}

				fmt.Printf("%s\n", resourceName)
			}
			return nil
		},
	}
	getResourceName.Flags().Bool("id", false, "print only the resource id number")

	getPath := &cobra.Command{
		Use:          "get-path [resource-name...]",
		Short:        "Get path of resource name",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			hierarchy, err := internal.LoadResourceHierarchy(ctx, nil)
			if err != nil {
				return err
			}
			for _, resourceName := range args {
				path, err := hierarchy.GetPathByResourceName(resourceName)
				if err != nil {
					return err
				}
				fmt.Printf("%s\n", path)
			}
			return nil
		},
	}

	root.AddCommand(ls)
	root.AddCommand(getResourceName)
	root.AddCommand(getPath)

	_ = root.Execute()
}
