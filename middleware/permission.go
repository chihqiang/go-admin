package middleware

import (
	"net/http"
	"strings"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
)

func Permission(skipRoutes ...string) httpx.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			for _, route := range skipRoutes {
				if strings.HasPrefix(r.RequestURI, route) {
					next(w, r)
					return
				}
			}

			account := AccountFromContext(r.Context())
			if account == nil {
				httpx.WriteHTTPError(w, httpx.CodeUnauthorized, "未登录")
				return
			}

			if !account.Status {
				httpx.WriteHTTPError(w, httpx.CodeUnauthorized, "账号已被禁用")
				return
			}

			if len(account.Roles) == 0 {
				httpx.WriteHTTPError(w, httpx.CodeForbidden, "无权限访问")
				return
			}

			method := r.Method
			uri := r.RequestURI
			if idx := strings.IndexByte(uri, '?'); idx > 0 {
				uri = uri[:idx]
			}

			seen := make(map[string]bool)
			for _, role := range account.Roles {
				for _, menu := range role.Menus {
					if menu.APIMethod == "" || menu.APIURL == "" {
						continue
					}
					key := menu.APIMethod + " " + menu.APIURL
					if seen[key] {
						continue
					}
					seen[key] = true
					if method == menu.APIMethod && strings.HasPrefix(uri, menu.APIURL) {
						next(w, r)
						return
					}
				}
			}

			logger.Warn("permission denied",
				logger.Int64("account_id", account.ID),
				logger.String("method", method),
				logger.String("uri", r.RequestURI),
			)

			httpx.WriteHTTPError(w, httpx.CodeForbidden, "无权限访问")
		}
	}
}
