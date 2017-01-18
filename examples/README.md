# Example tasks

[This](tasks/docker-task.json) example task will publish metrics to **file** 
from the mock plugin.  

## Running the example

### Requirements
 * `docker` and `docker-compose` are **installed** and **configured** 

### Collecting from docker in docker
Run the script `./run-docker-file.sh`.

### Collecting from your docker
Run the script `./run-dockerception.sh`. 

## Files

- [run-docker-file.sh](run-docker-file.sh) 
    - This script launchs docker in docker  
- [run-dockerception.sh](run-dockerception.sh) 
    - This script runs the example inside the Snap container
- [tasks/docker-task.json](tasks/docker-task.json)
    - Snap task definition
- [docker-compose.yml](docker-compose.yml)
    - A docker compose file which defines the "docker" container.
- [docker-file.sh](docker-file.sh)
    - Downloads `snapteld`, `snaptel`, `snap-plugin-publisher-file`,
    `snap-plugin-collector-docker` and starts the task 
    [tasks/docker-task.json](tasks/docker-task.json).
- [.setup.sh](.setup.sh)
    - Verifies dependencies and starts the containers.  It's called 
    by [run-docker-file.sh](run-docker-file.sh).