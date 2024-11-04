# Graph Pipeline

This job manager is a Go application designed to manage and schedule jobs concurrently. 
It supports both job queueing from various data sources like Weaviate and Macrostrat, or directly via POST requests. 
This system is structured to work with worker nodes that execute and store results of jobs.

## Details
- **Health Checks:** Regular health checks are performed to ensure jobs have not timed out. 
- **Data Loaders:** Supports different data loaders which can be specified at runtime. 

## Getting Started

### Running the Application

#### Running the Coordinator Individually

1. Set the necessary environment variables:
  - `GROUP_SIZE`: Defines the number of jobs processed in a batch. The manager will not queue a new group until the current one is completed.
  - `MINI_BATCH_SIZE`: Specifies the number of tasks in each job (ie. number of paragraphs in a job)
  - `HEALTH_INTERVAL`: Time interval in seconds for how often the manager checks for timed out jobs.
  - `HEALTH_TIMEOUT`: Duration in seconds after which a job is considered timed out if no health check is received. Jobs will be requeued if they have timed out.
  - `PIPELINE_ID`: The ID of the pipeline. 
  - `RUN_ID`: The ID of the current run.

2. Run the application:
  - Build the manager:
```docker build -t job-manager .```
  - Run the container in on-demand mode:
```docker run -it --env-file .env -p 50051:50051 -p 8000:8000 job-manager --set-loader=false```

    - Port `50051` is used for gRPC communication with worker nodes. Port `8000` is exposed for the `/submit_job` endpoint.
    - The port should match the address in the worker's `MANAGER_HOST` environment variable.

  - Optionally, specify a data loader when running the container (currently supports "weaviate" and "map_descrip"):
```docker run -it --env-file .env -p 50051:50051 job-manager --loader-type=weaviate```

    - This will continuously queue data from the specified data source.

#### Using Docker Compose to run with workers

1. Set necessary environment variables:
  - Initialize the environment variables listed above along with any needed for the worker container.
2. Change Docker Compose settings
  - Update the worker section of the `docker-compose.yml` file to be compatible with the worker container.
  - Change the coordinator entrypoint to set the data loader if desired.
3. Run the application
  - Run ```docker-compose up --build```

## API Endpoints

- `POST /submit_job`: Submit a new job.
  - Example body for weaviate data:
  ```json
  {
    "weaviate_data": {
      "paragraph_ids": [
        "0f8ce52f-8f0e-4b58-a6a6-7515a9965526",
        "53947580-833f-4eb3-8413-efbbddfa890b"
      ]
    }
  }
  ```
  - Example body for map description data:
  ```json
  {
    "map_description_data": {
      "legend_ids": [
        36046,
        36049
      ]
    }
  }
  ```
