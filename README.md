# Auto scaling framework, simulator and runtime

The repository contain an auto scaling framework, simulator and runtime
scaling accross multiple execution environments for an application
that submit jobs. These should be well defined jobs that execute and complete.
The framework is not designed to be used with continuously running jobs,
such as web servers.

## META-pipe
The framework was designed for the META-pipe backend and the implemented
estimator and simulator / runtime endpoints support only META-pipe jobs.

## Usage

To generate the database tables, use the provided "create_database.sql"
script.

The application can be launched in two modes, one for the simulator and
one for the runtime service. (Note that the runtime is not fully implemented).
The updateDB flag can be set at runtime to initialize the database with META-pipe jobs.
This should be done at least once to download the estimator training data.

## Dependencies
- gorilla/mux https://github.com/gorilla/mux
- lib/pq https://github.com/lib/pq
- segmentio/ksuid https://github.com/segmentio/ksuid
- sajari/regression https://github.com/sajari/regression

