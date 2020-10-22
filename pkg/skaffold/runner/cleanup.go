/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runner

import (
	"context"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/color"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/version"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/label"
	"io"
	"runtime"
	"time"
)

func (r *SkaffoldRunner) Cleanup(ctx context.Context, out io.Writer) error {
	return r.deployer.Cleanup(ctx, out)
}

func (r * SkaffoldRunner) Finalize(ctx context.Context, out io.Writer) {
	color.Blue.Fprintln(out, "Writing metrics...")
	meter := global.Meter("skaffold")
	opts := r.runCtx.Opts

	labels := []label.KeyValue{
		label.String("version", version.Get().Version),
		label.String("os", runtime.GOOS),
		label.String("arch", runtime.GOARCH),
		label.String("command", string(opts.Mode())),
		label.Float64("duration", time.Since(r.start).Seconds()),
	}

	switch opts.Mode() {
	case config.RunModes.Run:
		labels = append(labels, r.finalizeRun(ctx, out, meter, labels)...)
		break
	case config.RunModes.Build:
		r.finalizeBuild(ctx, out, meter, labels)
		break
	case config.RunModes.Dev:
		labels = append(labels, r.finalizeDev(ctx, out, meter, labels)...)
		break
	case config.RunModes.Render:
		r.finalizeRender(ctx, out, meter)
		break
	case config.RunModes.Deploy:
		labels = append(labels, r.finalizeDeploy(ctx, out, meter, labels)...)
		break
	case config.RunModes.Debug:
		labels = append(labels, r.finalizeDebug(ctx, out, meter, labels)...)
		break
	}

	runCounter := metric.Must(meter).NewInt64Counter("runs", metric.WithDescription("Skffold Invocations"))
	runCounter.Add(ctx, 1, labels...)

	commandCounter := metric.Must(meter).NewFloat64ValueRecorder(string(opts.Mode()),
		metric.WithDescription("durations of skaffold " + string(opts.Mode()) + " in seconds"))
	commandCounter.Record(ctx, time.Since(r.start).Seconds())
}

func (r * SkaffoldRunner) finalizeBuild(ctx context.Context, out io.Writer, meter metric.Meter, labels []label.KeyValue) {
	countBuilders(ctx, r.runCtx.Cfg.Build.Artifacts, meter, labels)
}

func (r * SkaffoldRunner) finalizeRender(ctx context.Context, out io.Writer, meter metric.Meter) {

}

func (r * SkaffoldRunner) finalizeDeploy(ctx context.Context, out io.Writer, meter metric.Meter, labels []label.KeyValue) []label.KeyValue {
	return countDeployer(ctx, r.runCtx.Cfg.Deploy.DeployType, meter)
}

func (r * SkaffoldRunner) finalizeDev(ctx context.Context, out io.Writer, meter metric.Meter, labels []label.KeyValue) [] label.KeyValue {
	countBuilders(ctx, r.runCtx.Cfg.Build.Artifacts, meter, labels)
	ls := countDeployer(ctx, r.runCtx.Cfg.Deploy.DeployType, meter)
	syncCounter := metric.Must(meter).NewInt64ValueRecorder("sync/session/count",
		metric.WithDescription("Number of Syncs in a session"))
	syncCounter.Record(ctx, int64(r.devIteration))
	return ls
}

func (r * SkaffoldRunner) finalizeDebug(ctx context.Context, out io.Writer, meter metric.Meter, labels []label.KeyValue) []label.KeyValue {
	countBuilders(ctx, r.runCtx.Cfg.Build.Artifacts, meter, labels)
	return countDeployer(ctx, r.runCtx.Cfg.Deploy.DeployType, meter)
}

func (r * SkaffoldRunner) finalizeRun(ctx context.Context, out io.Writer, meter metric.Meter, labels []label.KeyValue) []label.KeyValue {
	countBuilders(ctx, r.runCtx.Cfg.Build.Artifacts, meter, labels)
	return countDeployer(ctx, r.runCtx.Cfg.Deploy.DeployType, meter)
}

func countBuilders(ctx context.Context, artifacts []*latest.Artifact, meter metric.Meter, labels []label.KeyValue) {
	buildCounter := metric.Must(meter).NewInt64Counter("artifact/types",
		metric.WithDescription("Count of each artifact type"))

	builderToCountMap := make(map[string]int)

	for _, artifact := range artifacts {
		at := artifact.ArtifactType
		var artifactType string
		if at.DockerArtifact != nil {
			artifactType = "docker"
		}
		if at.BazelArtifact != nil {
			artifactType = "bazel"
		}
		if at.CustomArtifact != nil {
			artifactType = "custom"
		}
		if at.BuildpackArtifact != nil {
			artifactType = "buildpacks"
		}
		if at.JibArtifact != nil {
			artifactType = "jib"
		}
		if at.KanikoArtifact != nil {
			artifactType = "kaniko"
		}
		_, ok := builderToCountMap[artifactType]
		if !ok {
			builderToCountMap[artifactType] = 0
		}
		builderToCountMap[artifactType]++
	}

	for k, v := range builderToCountMap {
		var ls []label.KeyValue
		ls = append(ls, labels...)
		buildCounter.Add(ctx, int64(v), append(ls, label.String("builder", k))...)
	}
}

func countDeployer(ctx context.Context, deployType latest.DeployType, meter metric.Meter) []label.KeyValue {
	deployerCounter := metric.Must(meter).NewInt64Counter("deployers",
		metric.WithDescription("Count of deployers used"))

	var labels []label.KeyValue
	if deployType.HelmDeploy != nil {
		labels = append(labels, label.String("deployer", "helm"))
	}
	if deployType.KptDeploy != nil {
		labels = append(labels, label.String("deployer", "kpt"))
	}
	if deployType.KubectlDeploy != nil {
		labels = append(labels, label.String("deployer", "kubectl"))
	}
	if deployType.KustomizeDeploy != nil {
		labels = append(labels, label.String("deployer", "kustomize"))
	}

	deployerCounter.Add(ctx, 1, labels...)
	return labels
}