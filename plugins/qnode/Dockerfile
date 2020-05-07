FROM golang:1.14-alpine
RUN apk add --no-cache git gcc g++

# Get Env Variable From Docker Compose
ARG NODE_CONFIG
ARG NODE_PORT

# Set The Current Working Directory Inside The Container
RUN mkdir /goshimmer
WORKDIR /goshimmer

# Copy Everything From The Current Directory And Download Dependencies
COPY . .
RUN go mod download

# Copy Default Config File To Qnode Directory
COPY ./${NODE_CONFIG} ./config.json

# Change Directory
WORKDIR /goshimmer

EXPOSE ${NODE_PORT}

ENTRYPOINT ["go", "run", "./qnode-test/main.go"]
