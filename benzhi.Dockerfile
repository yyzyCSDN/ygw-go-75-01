FROM golang:1.23.12 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o /aquarcirc ./cmd/aquarcirc

FROM golang:1.23.12
ENV GOPROXY=off
ENV GOSUMDB=off
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
COPY --from=build /aquarcirc /usr/local/bin/aquarcirc
EXPOSE 8080
CMD ["/usr/local/bin/aquarcirc"]
