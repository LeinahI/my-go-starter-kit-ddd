package http

import (
	"database/sql"
	"net/http"

	"github.com/yourorg/ws/internal/transport/http/response"
)

// Health godoc
//
//	@Summary		Health check
//	@Description	Returns OK when the API and database are reachable.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	SwaggerHealthEnvelope
//	@Failure		503	{object}	SwaggerErrorEnvelope
//	@Router			/up [get]
func Health(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
