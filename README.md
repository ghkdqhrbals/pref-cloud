http benchmark server
=========
This is a simple http server that can be used to benchmark http clients.

## Usage

```bash
$ go run .
```

this will start a server on port 8080. You can change the port by setting the `PORT` environment variable.

## Example

* Request `POST /test` : This endpoint will return a 200 status code with a response body of `{"status": "ok"}`.

```bash
```json
{
  "url":"http://localhost:9001/board",
  "method":"POST",
  "numUsers":10,
  "numReqs":1000
}
```

* Response `POST /test` 200
> We serve the response in the following format: 

```json
{
  "ID": 26,
  "CreatedAt": "2024-02-27T21:30:21.618101+09:00",
  "UpdatedAt": "2024-02-27T21:30:21.618101+09:00",
  "DeletedAt": null,
  "url": "http://localhost:9001/board",
  "method": "POST",
  "total_requests": 10000,
  "total_errors": 0,
  "total_success": 10000,
  "StatusCodeCount": {
    "200": 10000
  },
  "total_users": 10,
  "total_duration": 1,
  "mttfb_average": "28.188ms",
  "MTTFBPercentiles": {
    "p50": "13.386ms",
    "p75": "23.908ms",
    "p90": "49.640ms",
    "p95": "103.998ms",
    "p99": "224.680ms"
  },
  "tps_average": 5883650.805324704,
  "TPSPercentiles": {
    "p50": 312,
    "p75": 113,
    "p90": 54,
    "p95": 23,
    "p99": 17
  }
}
```
