# OpenTelemetry (OTel) + SigNoz Tracing Demo

This project is a minimal demonstration of integrating **OpenTelemetry (OTel)** with **SigNoz** in a Go application.  
Currently, it focuses **only on the `trace` signal** for observability.

---

## 🚀 Requirements

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Postman](https://www.postman.com/) or `curl`

---

## 🛠️ How to Run

### 1. Start SigNoz Locally

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

---

### 2. Confirm Collector Port

Check the collector ports using:

```bash
docker ps
```

You should see something like this:

```
0.0.0.0:4317-4318->4317-4318/tcp
```

These ports are used for sending trace data via gRPC (4317) and HTTP (4318).

---

### 3. Run the Go App

Run the demo app locally:

```bash
go run .
```

Make a request to your API using Postman or curl. The traces will be automatically sent to SigNoz and appear in the dashboard.

---

## 📚 References

- [SigNoz sample Go app](https://github.com/SigNoz/sample-golang-app/blob/master/main.go)
- [OTel Go Collector Example](https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/otel-collector/main.go)
- [OpenTelemetry + Echo guide (Coroot)](https://docs.coroot.com/tracing/opentelemetry-go?http-server=echo)

---

## 📌 Notes

- Ensure the ports used by the collector are correctly mapped and accessible from your Go app.
- This project can be extended later to support other OTel signals like metrics and logs.