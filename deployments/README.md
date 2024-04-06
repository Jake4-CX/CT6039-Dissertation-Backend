Master:
 - `docker build -f deployments/master.dockerfile -t loadtest-master .`
 - `docker run -d -p 8080:8080 -p 15672:15672 -p 5671:5671 -p 5672:5672 --name master-container -v ${PWD}/.env:/app/.env loadtest-master`
Worker:
 - `docker build -f deployments/worker.dockerfile -t loadtest-worker .`
 - `docker run -d --name loadtest-worker -v ${PWD}/.env:/app/.env loadtest-worker