package auth

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

func NewCredential(clientID string) (*azidentity.InteractiveBrowserCredential, error) {
	c, err := cache.New(nil)
	if err != nil {
		return nil, err
	}

	return azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		ClientID: clientID,
		TenantID: "organizations",
		Cache:    c,
	})
}

func GetToken(ctx context.Context, cred *azidentity.InteractiveBrowserCredential) (string, error) {
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return "", err
	}
	return token.Token, nil
}
