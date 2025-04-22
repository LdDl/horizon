package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rest"
	"github.com/LdDl/horizon/rest/docs"
	"github.com/LdDl/horizon/rpc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"google.golang.org/grpc"
)

var (
	addrFlag   = flag.String("h", "0.0.0.0", "Bind address")
	portFlag   = flag.Int("p", 32800, "Port")
	fileFlag   = flag.String("f", "graph.csv", "Filename of *.csv file (you can get one using https://github.com/LdDl/osm2ch#osm2ch)")
	sigmaFlag  = flag.Float64("sigma", 50.0, "σ-parameter for evaluating emission probabilities")
	betaFlag   = flag.Float64("beta", 30.0, "β-parameter for evaluating transition probabilities")
	lonFlag    = flag.Float64("maplon", 0.0, "initial longitude of front-end map")
	latFlag    = flag.Float64("maplat", 0.0, "initial latitude of front-end map")
	zoomFlag   = flag.Float64("mapzoom", 1.0, "initial zoom of front-end map")
	apiPath    = "api"
	apiVersion = "0.1.0"

	grpcEnable   = flag.Bool("grpc", false, "Enable gRPC server")
	grpcAddrFlag = flag.String("gh", "0.0.0.0", "Bind gRPC address")
	grpcPortFlag = flag.Int("gp", 32801, "gRPC port")
	grpcReflect  = flag.Bool("gr", false, "Enable gRPC reflection")

	//go:embed index.html
	webPage string
)

// @title API for working with Horizon
// @version 0.1.0

// @contact.name API support
// @contact.url https://github.com/LdDl/horizon#table-of-contents
// @contact.email sexykdi@gmail.com

// @BasePath /

// @schemes http https
func main() {
	flag.Parse()

	// Init web page
	webPage = fmt.Sprintf(webPage, *lonFlag, *latFlag, *zoomFlag)

	// Init map matcher engine
	hmmParams := horizon.NewHmmProbabilities(*sigmaFlag, *betaFlag)
	matcher, err := horizon.NewMapMatcher(hmmParams, *fileFlag)
	if err != nil {
		fmt.Println(err)
		return
	}

	config := fiber.Config{
		DisableStartupMessage: false,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println("error:", err)
			return ctx.Status(500).JSON(map[string]string{"Error": "undefined"})
		},
		IdleTimeout: 10 * time.Second,
	}
	allCors := cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowHeaders:  "Origin, Authorization, Content-Type, Content-Length, Accept, Accept-Encoding, X-HttpRequest",
		AllowMethods:  "GET, POST, PUT, DELETE",
		ExposeHeaders: "Content-Length",
		MaxAge:        5600,
	})

	// Init REST API server
	server := fiber.New(config)
	server.Use(allCors)
	server.Get("/", rest.RenderPage(webPage))
	apiGroup := server.Group(apiPath)
	apiVersionGroup := apiGroup.Group(fmt.Sprintf("/v%s", apiVersion))

	apiVersionGroup.Post("/mapmatch", rest.MapMatch(matcher))
	apiVersionGroup.Post("/shortest", rest.FindSP(matcher))
	apiVersionGroup.Post("/isochrones", rest.FindIsochrones(matcher))

	docsStaticGroup := apiVersionGroup.Group("/docs")
	docsStaticGroup.Use("/", docs.PrepareStaticAssets())

	docsGroup := apiVersionGroup.Group("/docs")
	docsGroup.Use("/", docs.PrepareStaticPage())

	// Init gRPC API server if needed
	if *grpcEnable {
		grpcServer, err := rpc.NewMicroserice(matcher, *grpcReflect)
		if err != nil {
			fmt.Println("Can't prepare gRPC API instance", err)
			return
		}
		stdListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *grpcAddrFlag, *grpcPortFlag))
		if err != nil {
			fmt.Println("Can't prepare standart listener for gRPC API instance", err)
			return
		}
		errChan := make(chan error, 1)
		go func(s *grpc.Server, l net.Listener, errCh chan<- error) {
			fmt.Printf("Starting gRPC API  server on %s:%d\n", *grpcAddrFlag, *grpcPortFlag)
			fmt.Println("gRPC reflection enabled:", *grpcReflect)
			if err := s.Serve(l); err != nil {
				errCh <- fmt.Errorf("gRPC API server error: %w", err)
				return
			}
		}(grpcServer, stdListener, errChan)
		go func(errCh <-chan error) {
			err := <-errCh
			fmt.Println("Can't start gRPC API server", err)
			os.Exit(1)
		}(errChan)
	}

	// Start REST API server
	if err := server.Listen(fmt.Sprintf("%s:%d", *addrFlag, *portFlag)); err != nil {
		fmt.Println("Can't start REST API server", err)
		return
	}
}
