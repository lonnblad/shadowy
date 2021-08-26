# shadowy

## Setup

### Primary

- container #1
- shadowcar #1

### Secondary

- container #2
- shadowcar #2

### Candidate

- container #3
- shadowcar #3

### Shadowerse

-

## Shadowcar (primary)

- Incomming requests and outgoing responses are sent to shadowerse
- Outbound requests and inbound responses are sent to shadowerse

## Shadowcar (secondary)

- Outgoing responses are sent to shadowerse
- Outbound requests are sent to shadowerse

## Shadowcar (candidate)

- Outgoing responses are sent to shadowerse
- Outbound requests are sent to shadowerse

## Shadowerse

- Multiplex primary requests to secondary and candidate
- Diff responses between primary, secondary and candidate

## Paste-Bin

```
primary ---- insert --> postgres
primary ---- query ---> postgres

secondary -- insert --> postgres
secondary -- query ---> postgres

candidate -- insert --> postgres
candidate -- query ---> postgres

CD -> shadowerse

client -> primary sidecar -> primary service
                          -> shadowerse -> secondary sidecar -> secondary service
                                        -> candidate sidecar -> candidate service

secondary service -> secondary sidecar -> shadowerse
candidate service -> candidate sidecar -> shadowerse
primary service   -> primary sidecar   -> shadowerse
                                       -> client

shadowerse -> CD

mode=primary port=9011 service_port=8080 shadowerse_port=8090 go run cmd/shadowcar/main.go
port=9010 go run cmd/demo-service/main.go

mode=secondary port=9021 service_port=9020 shadowerse_port=8090 go run cmd/shadowcar/main.go
port=9020 go run cmd/demo-service/main.go

mode=candidate port=9031 service_port=9030 shadowerse_port=8090 go run cmd/shadowcar/main.go
port=9030 go run cmd/demo-service/main.go

port=8090 secondaryPort=9021 candidatePort=9031 go run cmd/shadowerse/main.go


primary service [9010]   -> primary sidecar [9011]   -> shadowerse [8090]
secondary service [9020] -> secondary sidecar [9021] -> shadowerse [8090]
candidate service [9030] -> candidate sidecar [9031] -> shadowerse [8090]


service
- port

sidecar
- mode
- port
- service_port
- shadowerse_port

shadowerse
- port
```
