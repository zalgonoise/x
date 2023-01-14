package main

import (
	"context"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/ghttp/example/server"
)

func main() {
	srv := server.New()
	srv.HTTP = ghttp.NewServer(8080, srv.Endpoints())

	err := srv.HTTP.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
