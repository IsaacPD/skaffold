package instrumentation

import (
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/mitchellh/go-homedir"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/resource"
	"path/filepath"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"os"

	"go.opentelemetry.io/otel/exporters/stdout"
)

type TraceHelper struct {
	Tracer trace.Tracer
	spanMap map[trace.Span]map[string]*label.Value
}

var helper *TraceHelper = nil

func InitBasicMetrics() *push.Controller {
	home, err := homedir.Dir()
	path := filepath.Join(home, constants.DefaultSkaffoldDir, constants.DefaultMetricFile)
	file, err := os.OpenFile(path, os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		logrus.Panicf("failed to create file for spans\n", err)
		return nil
	}

	pusher, err := stdout.InstallNewPipeline([]stdout.Option{
		stdout.WithQuantiles([]float64{0.5}),
		stdout.WithPrettyPrint(),
		stdout.WithWriter(file),
	}, nil)
	if err != nil {
		logrus.Panicf("failed to initialize metric stdout exporter %v", err)
	}
	return pusher
}

func InitBasicTracer() func() {
	var err error
	home, err := homedir.Dir()
	path := filepath.Join(home, constants.DefaultSkaffoldDir, constants.DefaultSpanFile)
	file, err := os.OpenFile(path, os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		logrus.Panicf("failed to create file for spans\n", err)
		return nil
	}
	exp, err := stdout.NewExporter(stdout.WithPrettyPrint(), stdout.WithWriter(file))
	if err != nil {
		logrus.Panicf("failed to initialize stdout exporter %v\n", err)
		return nil
	}
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(
			sdktrace.Config{
				DefaultSampler: sdktrace.AlwaysSample(),
			},
		),
		sdktrace.WithSpanProcessor(bsp),
	)
	global.SetTracerProvider(tp)
	return bsp.Shutdown
}

func InitTracer() func() {
	projectID := os.Getenv("PROJECT_ID")

	_, flush, err := texporter.InstallNewPipeline(
		[]texporter.Option{texporter.WithProjectID(projectID)},
		// For this example code we use sdktrace.AlwaysSample sampler to sample all traces.
		// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		logrus.Fatal(err)
	}
	return flush
}

func InitMeter() *push.Controller {
	projectID := os.Getenv("PROJECT_ID")

	pusher, err := mexporter.InstallNewPipeline(
		[]mexporter.Option{mexporter.WithProjectID(projectID)},
		push.WithResource(resource.New(
			label.String("installation_id", "isaacfakeinstallid"),
		)),
	)
	if err != nil {
		logrus.Fatalf("Failed to establish pipeline")
	}
	return pusher
}

func GetHelper() *TraceHelper {
	if helper == nil {
		helper = &TraceHelper{
			Tracer: global.Tracer("app/skaffold"),
			spanMap: make(map[trace.Span]map[string]*label.Value),
		}
	}
	return helper
}

func (th *TraceHelper) SetSpanAttribute(span trace.Span, key string, val label.Value) {
	if th.spanMap[span] == nil {
		th.spanMap[span] = make(map[string]*label.Value)
	}
	th.spanMap[span][key] = &val
}

func (th *TraceHelper) GetSpanAttribute(span trace.Span, key string) *label.Value {
	if th.spanMap[span] != nil {
		return th.spanMap[span][key]
	}
	return nil
}

func (th *TraceHelper) Finalize(span trace.Span) {
	if th.spanMap[span] != nil {
		for k, v := range th.spanMap[span] {
			span.SetAttributes(label.KeyValue{Key: label.Key(k), Value: *v})
		}
	}
	span.End()
}