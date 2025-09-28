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

1. Review the gRPC transport credentials, how would I grant TLS support for the communications?
