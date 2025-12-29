package main

import (
	_ "police-trafic-api-frontend-aligned/docs"
	"police-trafic-api-frontend-aligned/internal/app"

	"go.uber.org/fx"
)

// @title           Police Traffic API - Frontend Aligned
// @version         1.0
// @description     API backend alignée parfaitement avec le frontend
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@police-traffic.ci

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	fx.New(app.BuildApp()).Run()
}