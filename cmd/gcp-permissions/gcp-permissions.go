package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	resourcemanagerpb "cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
)

func main() {
	ctx := context.Background()

	// Command-line flags
	var orgId, folderId string
	flag.StringVar(&orgId, "org", "", "Organization ID (e.g., '123456789012')")
	flag.StringVar(&folderId, "folder", "", "Folder ID (e.g., '123456789012')")
	flag.Parse()

	if orgId == "" && folderId == "" {
		log.Fatalf("Either the Org ID or Folder ID must be specified using the -org or -folder flag")
	}

	client, err := resourcemanager.NewFoldersClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create folders client: %v", err)
	}
	defer client.Close()

	if folderId == "" {
		processOrg(ctx, client, orgId)
	} else {
		processFolderById(ctx, client, folderId)
	}
}

func processOrg(ctx context.Context, client *resourcemanager.FoldersClient, orgId string) {
	req := &resourcemanagerpb.ListFoldersRequest{
		Parent: fmt.Sprintf("organizations/%s", orgId),
	}
	it := client.ListFolders(ctx, req)
	for {
		folder, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Fatalf("Failed to list folders: %v", err)
		}

		processFolder(ctx, client, folder, folder.DisplayName)
	}
}

func processFolderById(ctx context.Context, client *resourcemanager.FoldersClient, folderId string) {
	req := &resourcemanagerpb.GetFolderRequest{
		Name: "folders/" + folderId,
	}
	folder, err := client.GetFolder(ctx, req)
	if err != nil {
		log.Fatalf("Failed to list folder %s - %v", folderId, err)
	}
	processFolder(ctx, client, folder, folderId)
}

func processFolder(ctx context.Context, client *resourcemanager.FoldersClient, folder *resourcemanagerpb.Folder, folderPath string) {
	folderID := folder.Name // Folder.Name is in the format "folders/123456789"
	folderDisplayName := folder.DisplayName

	header := fmt.Sprintf("Folder: %s / %s (ID: %s)\n", folderPath, folderDisplayName, folderID)

	printPolicies(ctx, client, folderID, header)

	req := &resourcemanagerpb.ListFoldersRequest{
		Parent: folderID,
	}

	it := client.ListFolders(ctx, req)
	for {
		subfolder, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Printf("Failed to list subfolders of %s: %v", folderID, err)
			break
		}

		newFolderPath := folderPath + " / " + subfolder.DisplayName
		processFolder(ctx, client, subfolder, newFolderPath)
	}
}

func printPolicies(ctx context.Context, client *resourcemanager.FoldersClient, folderID, header string) {
	iamReq := &iampb.GetIamPolicyRequest{
		Resource: folderID,
	}

	policy, err := client.GetIamPolicy(ctx, iamReq)
	if err != nil {
		log.Printf("Failed to get IAM policy for folder %s: %v\n", folderID, err)
		return
	}

	if len(policy.Bindings) == 0 {
		return
	}
	fmt.Println(header)

	for _, binding := range policy.Bindings {
		role := binding.Role
		for _, member := range binding.Members {
			fmt.Printf("\"%s\" = [ \"%s\"]\n", role, member)
		}
	}
	fmt.Println()
}
