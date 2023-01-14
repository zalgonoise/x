package main

import (
	"context"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/ghttp/example/server"
)

func main() {
	srv := server.New()
	srv.HTTP = ghttp.NewServer(srv.Endpoints(), 8080)

	err := srv.HTTP.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
