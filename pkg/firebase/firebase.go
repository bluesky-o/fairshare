package firebase

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type Client struct {
	Auth *auth.Client
}

func NewClient(serviceAccountPath string) (*Client, error) {
	ctx := context.Background()
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, serviceAccountPath)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("falied to initilize firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get firebase auth client: %w", err)
	}
	
	log.Println("firebase auth client initilized successfully")

	return &Client{Auth: authClient}, nil
}

func (c *Client) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := c.Auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid firebase token: %w", err)
	}

	return token, nil
}
