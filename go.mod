module message-bus-otel

go 1.15

replace go.opentelemetry.io/otel/label v0.17.0 => go.opentelemetry.io/otel/attribute v0.20.0

replace go.opentelemetry.io/otel/attribute/attribute v0.17.0 => go.opentelemetry.io/otel/attribute v0.20.0

require (
	github.com/nats-io/nats-server/v2 v2.9.23 // indirect
	github.com/nats-io/nats.go v1.28.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/otlp v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	google.golang.org/grpc v1.53.0
)
