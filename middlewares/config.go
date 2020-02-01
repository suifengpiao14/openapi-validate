package middlewares

import (
	"github.com/suifengpiao14/openapi-validate/config"
	"github.com/urfave/negroni"
)

var (
	app *negroni.Negroni

	// MiddlewareStack on pREST
	MiddlewareStack []negroni.Handler

	// BaseStack Middlewares
	BaseStack = []negroni.Handler{
		negroni.Handler(negroni.NewRecovery()),
		negroni.Handler(negroni.NewLogger()),
		// 验证请求参数
		//ValidationRequest(),
		//ValidateResponse(),
		HandlerSet(),
	}
)

func initApp() {
	if len(MiddlewareStack) == 0 {
		MiddlewareStack = append(MiddlewareStack, BaseStack...)

		if !config.ServerConfig.Debug && config.ServerConfig.EnableDefaultJWT {
			MiddlewareStack = append(MiddlewareStack, JwtMiddleware(config.ServerConfig.JWTKey, config.ServerConfig.JWTAlgo))
		}
		if config.ServerConfig.CORSAllowOrigin != nil {
			MiddlewareStack = append(MiddlewareStack, Cors(config.ServerConfig.CORSAllowOrigin, config.ServerConfig.CORSAllowHeaders))
		}
	}
	app = negroni.New(MiddlewareStack...)
}

// GetApp get negroni
func GetApp() *negroni.Negroni {
	if app == nil {
		initApp()
	}
	return app
}
