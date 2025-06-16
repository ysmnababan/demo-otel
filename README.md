# OpenTelemetry (OTel) + SigNoz Tracing Demo

This project is a minimal demonstration of integrating **OpenTelemetry (OTel)** with or without **SigNoz** in a Go application.  
Currently, it focuses **only on the `trace` signal** for observability.

---

## 🚀 Requirements

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Postman](https://www.postman.com/) or `curl`

---

## 🛠️ How to Run

### Option A — Use SigNoz Dashboard (Full Setup)

#### 1. Start SigNoz Locally

You can run SigNoz using Docker:

```bash
git clone -b main https://github.com/SigNoz/signoz.git
cd signoz/deploy/docker
docker compose up -d --remove-orphans
```

After the containers are running, open your browser and go to:

```
http://localhost:8080/
```

This is the SigNoz dashboard where traces will be visualized.

#### 2. Confirm Collector Port

Check the collector ports using:

```bash
docker ps
```

You should see something like:

```
0.0.0.0:4317-4318->4317-4318/tcp
```

These ports are used for sending trace data via gRPC (4317) and HTTP (4318).

#### 3. Run the Go App

Run the demo app locally:

```bash
go run .
```

Make a request to your API using Postman or curl. The traces will be automatically sent to SigNoz and appear in the dashboard.

---

### Option B — Use OTel Collector **Without** Observability Backend (Debug Only)

You can also test OTel without any observability backend like SigNoz, Jaeger, or Prometheus.

Simply skip the SigNoz installation step and instead:

1. Use the provided `docker-compose.yml` (containing only the OTel Collector).
2. Start the OTel Collector with:

```bash
docker compose up -d
```

This setup logs trace data to the console using the `debug` exporter.

#### 🔍 View Trace Output

To see the collected trace logs:

- Open Docker Desktop and view container logs, or
- Use the CLI:

```bash
docker compose logs -f otel-collector
```

You should see output like:

```
2025-06-16T11:15:47.913Z  info  Traces  {
  "service.name": "otelcol",
  "otelcol.component.kind": "exporter",
  "otelcol.signal": "traces",
  "resource spans": 1,
  "spans": 3
}

Resource attributes:
-> library.language: Str(go)
-> service.name: Str(demo_otel)

Span #0
Trace ID : f81f9fe1941ebe2526eada0bc88af800
Parent ID : 3f8467a1665ac2dc
ID : a505640321935c51
Name : hello-world
Kind : Internal
Start time : 2025-06-16 11:15:48.8160842 +0000 UTC
End time   : 2025-06-16 11:15:48.8160842 +0000 UTC
```

This is useful for testing trace generation **without setting up a full observability stack**.

---

## 📚 References

- [SigNoz sample Go app](https://github.com/SigNoz/sample-golang-app/blob/master/main.go)
- [OTel Go Collector Example](https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/otel-collector/main.go)
- [OpenTelemetry + Echo guide (Coroot)](https://docs.coroot.com/tracing/opentelemetry-go?http-server=echo)

---

## 📌 Notes

- Ensure the ports used by the collector (`4317`, `4318`) are correctly mapped and accessible from your Go app.
- This project currently only supports traces but can be extended to include metrics and logs.
- Using the debug exporter is perfect for local development or integration testing without third-party UIs.
