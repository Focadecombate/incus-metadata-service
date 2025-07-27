package internal_routes

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
)

type Handler struct {
	Config   *config.Config
	Database db.Querier
}
