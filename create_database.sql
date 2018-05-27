CREATE TABLE IF NOT EXISTS estimator_training
(
  runtime       BIGINT,
  tag           VARCHAR(255),
  jobid         VARCHAR(255) NOT NULL,
  datasetsize   INTEGER,
  queueduration BIGINT
);

CREATE UNIQUE INDEX IF NOT EXISTS estimator_training_jobid_pk
  ON estimator_training (jobid);

CREATE UNIQUE INDEX IF NOT EXISTS estimator_training_jobid_uindex
  ON estimator_training (jobid);

ALTER TABLE estimator_training
  ADD CONSTRAINT estimator_training_jobid_pk
PRIMARY KEY (jobid);

CREATE TABLE IF NOT EXISTS metapipe_parameters
(
  inputcontigscutoff     INTEGER,
  useblastuniref50       BOOLEAN,
  useinterproscan5       BOOLEAN,
  usepriam               BOOLEAN,
  removenoncompletegenes BOOLEAN,
  exportmergedgenbank    BOOLEAN,
  useblastmarref         BOOLEAN,
  jobid                  VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS metapipe_parameters_jobid_pk
  ON metapipe_parameters (jobid);

CREATE UNIQUE INDEX IF NOT EXISTS metapipe_parameters_jobid_uindex
  ON metapipe_parameters (jobid);

ALTER TABLE metapipe_parameters
  ADD CONSTRAINT metapipe_parameters_jobid_pk
PRIMARY KEY (jobid);

ALTER TABLE metapipe_parameters
  ADD CONSTRAINT metapipe_parameters_estimator_training_jobid_fk
FOREIGN KEY (jobid) REFERENCES estimator_training;

CREATE TABLE IF NOT EXISTS cloud_events
(
  id             SERIAL NOT NULL,
  run_name       VARCHAR(255),
  created        TIMESTAMP,
  instance_id    VARCHAR(255),
  type           VARCHAR(255),
  instance_type  VARCHAR(255),
  price          DOUBLE PRECISION,
  instance_state VARCHAR(255),
  cloud_name     VARCHAR(255)
);

CREATE UNIQUE INDEX IF NOT EXISTS cloud_events_id_pk
  ON cloud_events (id);

CREATE UNIQUE INDEX IF NOT EXISTS cloud_events_id_uindex
  ON cloud_events (id);

ALTER TABLE cloud_events
  ADD CONSTRAINT cloud_events_id_pk
PRIMARY KEY (id);

CREATE TABLE IF NOT EXISTS autoscaling_run
(
  id       SERIAL       NOT NULL,
  name     VARCHAR(255) NOT NULL,
  started  TIMESTAMP,
  finished TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS autoscaling_run_id_pk
  ON autoscaling_run (id);

CREATE UNIQUE INDEX IF NOT EXISTS autoscaling_run_id_uindex
  ON autoscaling_run (id);

CREATE UNIQUE INDEX IF NOT EXISTS autoscaling_run_name_uindex
  ON autoscaling_run (name);

ALTER TABLE autoscaling_run
  ADD CONSTRAINT autoscaling_run_id_pk
PRIMARY KEY (id);

ALTER TABLE cloud_events
  ADD CONSTRAINT cloud_events_autoscaling_run_name_fk
FOREIGN KEY (run_name) REFERENCES autoscaling_run (name);

CREATE TABLE IF NOT EXISTS simulator_events
(
  id             SERIAL NOT NULL,
  run_name       VARCHAR(255),
  queue_duration INTEGER,
  alg_timestamp  TIMESTAMP,
  tag            VARCHAR(255),
  cost_before    DOUBLE PRECISION,
  cost_after     DOUBLE PRECISION
);

CREATE UNIQUE INDEX IF NOT EXISTS simulator_events_id_uindex
  ON simulator_events (id);

CREATE UNIQUE INDEX IF NOT EXISTS simulator_events_pkey
  ON simulator_events (id);

ALTER TABLE simulator_events
  ADD CONSTRAINT simulator_events_pkey
PRIMARY KEY (id);

ALTER TABLE simulator_events
  ADD CONSTRAINT simulator_events_autoscaling_run_name_fk
FOREIGN KEY (run_name) REFERENCES autoscaling_run (name);

CREATE TABLE IF NOT EXISTS algorithm_job
(
  id            SERIAL NOT NULL,
  run_name      VARCHAR(255),
  jobid         VARCHAR(255),
  created       TIMESTAMP,
  started       TIMESTAMP,
  executiontime BIGINT,
  tag           VARCHAR(255),
  deadline      TIMESTAMP,
  priority      INTEGER,
  state         VARCHAR(255)
);

CREATE UNIQUE INDEX IF NOT EXISTS algorithm_job_id_uindex
  ON algorithm_job (id);

CREATE UNIQUE INDEX IF NOT EXISTS algorithm_job_pkey
  ON algorithm_job (id);

ALTER TABLE algorithm_job
  ADD CONSTRAINT algorithm_job_pkey
PRIMARY KEY (id);

ALTER TABLE algorithm_job
  ADD CONSTRAINT algorithm_job_autoscaling_run_name_fk
FOREIGN KEY (run_name) REFERENCES autoscaling_run (name);


