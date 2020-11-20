---
title: "Telemetry"
linkTitle: "Telemetry"
weight: 10
aliases: [/docs/metrics]
---

To help improve the quality of this product, we collect anonymized usage data. The breakdown of data we collect is as follows
- Exit Code
  - When skaffold finishes executing it returns an exit code which is collected and tracked
- Build Artifacts
  - The number of artifacts built in the current execution as defined in skaffold.yaml
- Builders
  - All the builders used to build the artifacts built
- Command
  - The command that is used to execute skaffold `dev, build, render, run, etc.`
- Version
  - The version of skaffold being used "v1.18.0, v1.19.1, etc."
- OS
  - The OS running skaffold as returned from runtime.GOOS
- Arch
  -  The OS running skaffold as returned from runtime.GOARCH
- Deployers
  - All the deployers used in the skaffold execution
- Duration
  - How long skaffold took to finish executing
- Error Code
  - Skaffold reports [error codes](/docs/references/api/grpc/#statuscode) and these are monitored in order to determine the most frequent errors
- Enum Flags
  - Any flags passed into Skaffld that have a pre-defined list of valid values e.g. "--cache-artifacts=false", "--mute-logs=["build", "deploy"ss]"
- Platform Type
  - Where skaffold is deploying to (sync, build, or google cloud build)
- Sync Type
  - The sync type used in the build configuration: infer, auto, and/or manual
- Dev Iterations
  - The error results of the various dev iterations and the reasons they were triggered. The triggers can be one sync, build or deploy.

This data is handled in accordance with our privacy policy <https://policies.google.com/privacy>.

You may choose to opt out of this collection by running the following command:

       skaffold config set collect-metrics false
