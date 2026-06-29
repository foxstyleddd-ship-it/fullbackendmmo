# Image autonome pour Railway / Fly.io / Cloud Run / n'importe quel host Docker.
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /gta6 .

FROM alpine:3.20
COPY --from=build /gta6 /gta6
ENV PORT=8080
EXPOSE 8080
CMD ["/gta6"]
