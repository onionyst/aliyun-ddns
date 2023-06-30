FROM golang:1.20-alpine AS build

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o /bin/ddns ./cmd/ddns/*

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=build /bin/ddns ./ddns

CMD [ "./ddns" ]
