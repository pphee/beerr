package databases

import (
	"context"
	"fmt"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"pok92deng/config"
)

func ConnectZitadel(ctx context.Context, cfg config.IZitadelConfig) (*client.Client, error) {
	api, err := client.New(ctx, zitadel.New(cfg.Domain()),
		client.WithAuth(client.DefaultServiceUserAuthentication(cfg.Key(), "openid", "urn:zitadel:iam:org:project:id:your_project_id:aud")),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating API client: %w", err)
	}
	return api, nil
}
