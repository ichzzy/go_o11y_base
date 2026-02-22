package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OtelConfig struct {
	Enabled     bool
	Endpoint    string  // OTLP gRPC Endpoint (e.g., localhost:4317)
	ServiceName string  // 服務名稱
	Env         string  // 部署環境 (local, dev, prod)
	SampleRatio float64 // 採樣率 (0.0 - 1.0)
}

type OTEL struct {
	tp   *sdktrace.TracerProvider
	conn *grpc.ClientConn
}

func InitOTEL(ctx context.Context, conf *OtelConfig) (*OTEL, error) {
	if !conf.Enabled {
		return &OTEL{}, nil
	}
	if conf.Endpoint == "" {
		return nil, fmt.Errorf("otel endpoint is empty")
	}

	// 建立 Resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(conf.ServiceName),
			semconv.DeploymentEnvironmentKey.String(conf.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 建立 gRPC 連線
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	conn, err := grpc.NewClient(conf.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}

	// 建立 Trace Exporter，負責將 Span 轉換成 OTLP 格式並透過 gRPC 送出
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// 配置 TracerProvider，套用 BatchSpanProcessor 來批量送出 Spans 給 Exporter
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(conf.SampleRatio)), // 採樣
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// 註冊全局 Provider, 以便 otelgin 等函式庫可以透過 otel.GetTracerProvider() 拿到單例
	otel.SetTracerProvider(tracerProvider)
	// 設定 Context 傳播機制, 可配置 baggage 到 context 中，讓 server 端獲取
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return &OTEL{
		tp:   tracerProvider,
		conn: conn,
	}, nil
}

// Shutdown 執行優雅關閉：先關閉 TracerProvider (確保 Spans 送出)，再關閉 gRPC 連線
func (o *OTEL) Shutdown(ctx context.Context) error {
	if o.tp != nil {
		// Shutdown 會強制刷新尚未送出的 Spans，並停止 SpanProcessors
		if err := o.tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("tracer provider shutdown failed: %w", err)
		}
	}
	if o.conn != nil {
		if err := o.conn.Close(); err != nil {
			return fmt.Errorf("grpc connection close failed: %w", err)
		}
	}
	return nil
}
