package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

func initTracer() func() {
	// Fetch the necessary settings (from environment variables, in this example).
	// You can find the API key via https://ui.honeycomb.io/account after signing up for Honeycomb.
	apikey, _ := os.LookupEnv("HONEYCOMB_API_KEY")
	dataset, _ := os.LookupEnv("HONEYCOMB_DATASET")
	var serviceName string
	if serviceName = os.Getenv("SERVICE_NAME"); serviceName == "" {
		serviceName = "message-bus-tracer"
	}

	// Initialize an OTLP exporter over gRPC and point it to Honeycomb.
	ctx := context.Background()
	exporter, err := otlp.NewExporter(
		ctx,
		otlpgrpc.NewDriver(
			otlpgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
			otlpgrpc.WithEndpoint("api.honeycomb.io:443"),
			otlpgrpc.WithHeaders(map[string]string{
				"x-honeycomb-team":    apikey,
				"x-honeycomb-dataset": dataset,
			}),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Configure the OTel tracer provider.
	resource := resource.Merge(resource.Default(), resource.NewWithAttributes(attribute.String("service.name", serviceName)))
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(provider)

	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	// This callback will ensure all spans get flushed before the program exits.
	return func() {
		ctx := context.Background()
		err := provider.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	cleanup := initTracer()
	defer cleanup()
	// the rest of your initialization...

	log.Println("Connecting to default url", nats.DefaultURL)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	tp := otel.GetTracerProvider()
	// messaging system specification https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/messaging.md
	// This should definitely be in a common library

	wait := make(chan struct{})
	goChan := make(chan struct{})
	go func(goChan, wait chan struct{}) {
		log.Println("Subscribed to message-bus")
		sub, err := nc.SubscribeSync("message-bus")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		close(goChan)
		log.Println("Waiting for message")
		msg, err := sub.NextMsg(10 * time.Second)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		log.Println("got message")

		ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(msg.Header))
		tracer := tp.Tracer("message-bus-otel-receive")
		ctx, span := tracer.Start(ctx, "NATS Receive")
		span.End()
		time.Sleep(100 * time.Millisecond)
		ctx, span = tracer.Start(ctx, "NATS Process")
		span.End()
		close(wait)
	}(goChan, wait)
	// Because it's both sync and async it's probably a bit lossy
	<-goChan
	tracer := tp.Tracer("message-bus-otel")
	ctx, span := tracer.Start(context.Background(), "NATS Send")
	span.SetAttributes(
		attribute.String("nats.url", nc.ConnectedUrl()),
		attribute.String("nats.addr", nc.ConnectedAddr()),
		attribute.String("messaging.system", "nats.io"),
		attribute.String("nats.subject", "message-bus"),
	)

	header := nats.Header{}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
	msg := &nats.Msg{
		Subject: "message-bus",
		Header:  header,
		Data:    []byte("This is a message with trace information embedded"),
	}
	log.Println("Publishing Message")
	nc.PublishMsg(msg)
	span.End()
	<-wait
	nc.Close()
}
