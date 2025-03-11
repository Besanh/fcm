<div align="center">
<h1>FCM API Microservices</h1>
</div>

# Introduction
Main functionlities of this product are connect client to FCM and push notification to mobile(ios, android) or website.

There are 3 parts, including main fcm process, worker background, logging and monitoring:

1. **fcm**: the heart of this product. It take on responsible for authentication and authorization, connect to **Goole FCM** and other functionalities.
2. **worker background**: take on responsible for getting messages from message queue, handling send and store logs into **MongoDB**
3. **logging and monitoring**: get logs from **MongoDB**, handling reports and statistics

# Roadmap
1. **Phase 1**: design and config the system architecture, build the structure of source code folder.
Design and develop the authentication use OAuth 2.0
2. **Phase 2**: develope apis push notification to FCM, process data in background queue as worker
3. **Phase 3**: develope logging and monitoring, implement swagger and other and other relevant technologies

# Architecture
- Programing Language: Golang(version 1.22 or above)
- Protocol: HTTP(REST API)
- NoSQL: MongoDB
- Caching: Redis
- Message Queue: Redis pub/sub or NATS Jetstream
- Design pattern: Repository pattern, Circuit breaker pattern
- Monitor api: Sentry, Jaeger, Opentelemetry
- Document: Swagger

The architecture mechanism

![Project Architecture](/assets/images/architecture.png)

# Project structure
1. **apis**: defines paths of apis
2. **repositories**: defines mongodb connection
3. **services**: defines logics
4. **common**: defines supported packages
5. **models**: defines structs
6. **servers**: defines server to run api
7. **tmp**: includes file log of this project
8. **build**: bash to build source code
9. **Dockerfile**: a Docker file help you to build an image
10. **assets**: includes images and media
11. **pkgs**: defines all needed libraries
12. **dockercomposes**: includes all docker compose to install and run docker container on server

# Commit guidelines
- `{action}-{write briefly content (max 50 characters)}`

# References
- Golang: https://golang.google.cn

- ORM: [Bun](https://bun.uptrace.dev/)
    - Link: https://bun.uptrace.dev
    - Git: https://github.com/uptrace/bun

- Go Generics:
    - Link: https://go.dev/doc/tutorial/generics

- A few libraries:
    - Gin: web server framework
    - OpenTelemetry: tracing service and request
    - Cobra: it's helps convert golang to command lines

# Setting FCM
I'm using FCM v1 API, accessing to fcm website and downloading file **<name>.json**
1. Open terminal and run the command
```
cat service-account.json | base64
```

2. Put **SECRET_KEY** value to file **.env. It's helps extreamly securing source code

3. Copy the outcome and put it to **FCM_CREDENTIAL_BASE64** in file **.env**

# Monitor
You must install **Jaeger** and **Opentelemetry** on the server and access below address to monitor
```
{{your IP or domain}}:16686
```


<div align="center">
Copyright © 2025 AnhLe. All rights reserved.
</div>