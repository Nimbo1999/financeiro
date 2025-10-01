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

...no items in pending list

### TODOS:

1. Verify the possibility of creating a finance app RabbitMQ Client class to use accross
   the different services of the finance platform. The goal would be extracting the RabbitMQ
   implementation that was build on the authentication service
   authentication/internal/messaging/\*.go and move to a module where the applications can
   reuse the client without having to implement the logic internally.
2. Refactor the Circuit Breaker pattern of each service. The goal is to create a Circuit Breaker
   module and make it available to all the services that would like to install it. This way I can
   have the benefit of a centralized implementation that would provide the same behavior and develop
   experience across the different services of my solution.
