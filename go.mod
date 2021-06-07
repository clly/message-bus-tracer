module message-bus-otel

go 1.15

replace go.opentelemetry.io/otel/label v0.17.0 => go.opentelemetry.io/otel/attribute v0.20.0

replace go.opentelemetry.io/otel/attribute/attribute v0.17.0 => go.opentelemetry.io/otel/attribute v0.20.0

require (
	github.com/nats-io/go-nats-examples v0.0.0-20190628222711-def6f82f468c // indirect
	github.com/nats-io/nats.go v1.11.0
	github.com/nats-io/stan.go v0.9.0 // indirect
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/otlp v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	google.golang.org/grpc v1.38.0
)
