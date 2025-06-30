# OpenTelemetry (OTel) + SigNoz Tracing Demo

This project is a minimal demonstration of integrating **OpenTelemetry (OTel)** with or without **SigNoz** in a Go application.  
Currently, it focuses **only on the `trace`and `log` signal** for observability.

---

## 🚀 Requirements

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Postman](https://www.postman.com/) or `curl`

---

## 🛠️ How to Run

### Option A — Use SigNoz Dashboard (Full Setup)
```
golang
  |
  =====(trace and metric signal)============> otel collector => Signoz
  |                                                 ^
  |                                                 | 
  =====(log signal)==========> docker stdout => fluentbit
```
#### 1. Start SigNoz Locally

You can start SigNoz using Docker:

```bash
git clone -b main https://github.com/SigNoz/signoz.git
cd signoz/deploy/docker
docker compose up -d --remove-orphans
```

Once the containers are running, open your browser and navigate to:

```
http://localhost:8080/
```

This is the SigNoz dashboard where your traces will be visualized.

> **Note:**  
> On your first visit, you'll be prompted to create and log in with a new user account. Make sure to remember your credentials.
>
> If you forget your login details, you can reset SigNoz by stopping the containers and removing the associated volumes (this will erase all data, including users):
>
> ```bash
> docker compose down -v
> docker compose up -d --remove-orphans
> ```
>
> After this, you can access the dashboard again and set up a new user.

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
#### 3. Create and Connect to a Docker Network

To ensure your Go app and the SigNoz OTel Collector can communicate, create a dedicated Docker network and connect the collector container to it:

```bash
docker network create otel-network
docker network connect otel-network signoz-otel-collector
```

This step allows your local Go application to send trace data to the collector using the network alias `signoz-otel-collector`.

#### 4. Run the Go App

Run the demo app in docker:


1. Navigate to the `with_signoz` folder.
2. Start the FluentBit and the Go App:

  ```bash
  docker compose up --build -d
  ```

3. Trigger the API endpoints:

  - [http://localhost:1323/](http://localhost:1323/)
  - [http://localhost:1323/for-loop](http://localhost:1323/for-loop)

---

### Option B — Use OTel Collector **Without** Observability Backend (Debug Only)

You can also experiment with OpenTelemetry **without connecting to any observability backend** (such as SigNoz, Jaeger, or Prometheus).

This setup allows you to see trace and log signals directly in your local environment:

```
golang
  |
  =====(trace signal)============> otel collector => collector_stdout
  |                                         ^
  |                                         | 
  =====(log signal)===> docker stdout => fluentbit
```

**Steps:**

1. Navigate to the `without_signoz` folder.
2. Start the OTel Collector and supporting containers:

  ```bash
  docker compose up --build -d
  ```

3. Trigger the API endpoints:

  - [http://localhost:1323/](http://localhost:1323/)
  - [http://localhost:1323/for-loop](http://localhost:1323/for-loop)

Trace data will be output to the console via the collector's `debug` exporter, and logs will be available in Docker stdout (optionally processed by Fluent Bit). This approach is ideal for local development and debugging without a full observability stack.

#### 🔍 View Trace And Log Output

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
- This project currently only supports traces and logs but can be easily extended to include metrics.
- Using the debug exporter is perfect for local development or integration testing without third-party UIs.
