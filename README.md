### Common commands

#### Create Migration file

```(bash)
# -seq
#   Use sequential numbers instead of timestamp
# -dir
#   Directory to place the file in
# -ext
#   File extension
# [NAME]
#   Migration name that would be added to the sql file
migrate create -seq -dir migrations -ext sql [NAME]
```

### Researchs:

1. Read and understand more about CircuitBreaker pattern in golang.

### TODOS:

1. Generate documentations on how the finance project handles the following topics:

   a. Project design principles and code pattern documentation;

   b. What is the testing pattern of the project;

   c. How each service should implement the gRPC connection, the gRPC client is exposed by
   each package on their /pkg folder.

2. Verify the possibility of creating a finance app RabbitMQ Client class to use accross
   the different services of the finance platform. The goal would be extracting the RabbitMQ
   implementation that was build on the authentication service
   authentication/internal/messaging/\*.go and move to a module where the applications can
   reuse the client without having to implement the logic internally.
3. Review the gRPC transport credentials, how would I grant TLS support for the communications?
