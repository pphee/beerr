package auth

import (
	"context"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"

	"golang.org/x/exp/slog"
	"pok92deng/config"
)

//var (
//	domain = flag.String("domain", "https://auth.cloudsoft.co.th", "your ZITADEL instance domain")
//	key    = flag.String("key", "/Users/p/Goland/pok92deng/254849895242924555.json", "path to your key.json")
//)

func InitZitadelAuthorization(ctx context.Context, cfg config.IZitadelConfig) (*authorization.Authorizer[*oauth.IntrospectionContext], error) {

	authZ, err := authorization.New(ctx, zitadel.New(cfg.Domain()), oauth.DefaultAuthorization(cfg.Key()))
	if err != nil {
		slog.Error("zitadel sdk could not initialize", "error", err)
		return nil, err
	}

	return authZ, nil
}
