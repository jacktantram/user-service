# User Service


## Repository
The repository follows a monorepo style approach.
It is structured as follows:

* `/github` - github setup/github actions
* `/build` - generated files such as protobuf definitions
* `/docs` - repo related docs containing data model & sequence diagrams.
* `/pkg` - shared libraries for project
* `/proto` - global protobuf definitions
* `/services` - containing any microservices

### Services
All services in `/service` directory should be containerised using Docker.
In order to spin up infrastructure run `make run`.

### Makefile
* `make lint` - run linter across project
* `make proto-generate`-protobuf generation
* `make build` - build containers
* `make run` - to run services
* `make test` - executes tests
* `make test-integration` - executes integration tests

### Protobuf
In order to generate definitions run `make proto-generate`. This will
generate code for languages specified in `buf.gen.yaml` based on
proto definitions in `/proto`. For extra language support
add in `buf.gen.yaml`


## Improvements 
* Depending on scale (CQRS)...
* Protobuf
    * CI - Responsible for protobuf generation to ensure no compatibility/versioning issues across machines.
    * `WIRE_JSON` - In order to share the protobuf schemas and avoid duplication I added the WIRE_JSON check. This was
      to avoid writing extra mapping functions. However, by doing this it stops the rpc/internal formats to be able to
      benefit from WIRE changes.
* Database
    * If more time would have written table driven tests to tidy up tests.
* Metrics
    * To add metrics I would look at adding [promhttp](https://github.com/prometheus/client_golang/tree/master/prometheus/promhttp) to be able to instrument the HTTP handler. This would enable
      dashboards to be built to track things like latency and number of requests
* Idempotency
    * Ideally an idempotency mechanism would be implemented to prevent clients making duplicate requests. It would also
      cache results of previous calls.
* Testing
  * Improve service layer tests, ran out of time to cover further edge cases
  * e2e tests
    * I would like to write e2e tests to spin up the gateway and call each endpoint validating that they work. This would be done by spinning up via docker-compose and writing BDD styled tests. The Ginkgo library is good for this.
* Production Readiness
  * All committed secrets, setup should be removed and injected in as a separate process.
  * In order to scale services accordingly Kubernetes could be used for each service
    so that they can be scaled independently and horizontally.
  * Monitors setup to track service health such as `/health`, `/metrics` endpoint as well
    as business metrics where alerts can be triggered.
