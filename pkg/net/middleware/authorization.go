package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/BEOpenSourceCollabs/EventManagementCore/pkg/net/constants"
	"github.com/BEOpenSourceCollabs/EventManagementCore/pkg/utils"
)

func ProtectMiddleware(next http.Handler, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//extract auth header
		authorization := r.Header.Get("Authorization")

		//return 401 if no authorization header
		if authorization == "" {
			utils.WriteErrorJsonResponse(w, constants.ErrorCodes.AuthInvalidAuthHeader, http.StatusUnauthorized, nil)
			return
		}

		//check if auth scheme is `Bearer` and token is present
		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.WriteErrorJsonResponse(w, constants.ErrorCodes.AuthInvalidAuthHeader, http.StatusUnauthorized, nil)
			return
		}

		//validate token
		token := parts[1]
		payload, err := utils.ValidateToken(token, secret)

		if err != nil {
			utils.WriteErrorJsonResponse(w, constants.ErrorCodes.AuthInvalidAuthToken, http.StatusUnauthorized, nil)
			return
		}

		ctx := context.WithValue(r.Context(), constants.USER_CONTEXT_KEY, payload)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
