FROM golang:1.13-alpine
RUN apk add --no-cache git gcc g++

# Get Env Variable From Docker Compose
ARG NODE_CONFIG
ARG NODE_PORT

# Set The Current Working Directory Inside The Container
RUN mkdir /app
WORKDIR /app

# Copy Everything From The Current Directory And Download Dependencies
COPY . .
RUN go get -u github.com/iotaledger/goshimmer
RUN go get -u .

# Copy Default Config File To Qnode Directory
COPY ./${NODE_CONFIG} ./plugins/qnode/qnode-test/config.json

EXPOSE ${NODE_PORT}

WORKDIR /app/qnode-test

ENTRYPOINT ["go", "run", "./main.go"]
