# Distributed Load Testing Tool - Backend

## Introduction

This repository houses the backend components of the distributed load testing tool designed to simplify advanced performance testing. The backend is developed using the Go programming language, following a master-slave architecture to efficiently manage load testing operations.

## Technology Stack

- **Go**: Primary programming language.
- **RabbitMQ**: For communication between backend nodes.
- **REST API**: For interfacing with the front end.
- **Socket.IO**: For real-time communication with the front end.
- **Docker**: For containerizing the backend services.
- **Kubernetes**: For orchestrating the containers, hosted on Digital Ocean.

## Architecture

The backend adopts a master-slave architecture where:

- The **Master** hosts a REST API and a Socket.IO server, facilitating communication with the front end and managing the slave nodes.
- **Slave nodes** perform the load tests as directed by the master, with coordination handled via RabbitMQ.

## Installation and Setup

To set up the backend using Docker:

1. Ensure Docker and Docker Compose are installed on your machine.
2. Clone the repository:

```bash
git clone https://github.com/Jake4-CX/CT6039-Dissertation-Backend.git
```

3. Navigate to the project directory:

```bash
cd load-test-backend
```

4. Use Docker Compose to start the services:

```bash
docker-compose up -d
```

This will pull the latest images for the master and worker nodes from `jake4/loadtest-master:latest` and `jake4/loadtest-worker:latest` and start the necessary services as defined in the `docker-compose.yml`.

## Usage

Once the services are running, the master node will be available at `http://localhost:8080`, providing access to the REST API and the Socket.IO server. The backend is also set up to communicate with slave nodes through RabbitMQ for distributing tasks.

## Deployment

For deploying the backend to a production environment:

- Configure the Kubernetes files under `/deployments/k8s` to suit your deployment needs.
- Use the provided manifests to deploy the backend services on a Kubernetes cluster, such as Digital Ocean's Kubernetes service.

## Backend Repository

For the backend source code and setup instructions, please visit the [frontend repository](https://github.com/Jake4-CX/CT6039-Dissertation-Front-End/).

## Contributing

Contributions are highly welcome. Please fork the repository and submit pull requests for any features or bug fixes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **Go Community**: For resources and support in backend development.
- **Digital Ocean**: For providing the infrastructure to host our services.
