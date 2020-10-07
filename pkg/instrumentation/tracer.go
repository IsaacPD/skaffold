package instrumentation

import (
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"os"
	"github.com/sirupsen/logrus"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type TraceHelper struct {
	Tracer trace.Tracer
}

var helper *TraceHelper = nil

func InitTracer() func() {
	projectID := os.Getenv("PROJECT_ID")

	// Create Google Cloud Trace exporter to be able to retrieve
	// the collected spans.
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

func GetHelper() *TraceHelper {
	if helper == nil {
		helper = &TraceHelper{
			Tracer: global.Tracer("app/skaffold"),
		}
	}
	return helper
}

