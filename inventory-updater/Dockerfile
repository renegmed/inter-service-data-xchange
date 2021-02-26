FROM golang:1.15.7-alpine as builder

WORKDIR /app

COPY main.go .
COPY go.mod .
COPY go.sum .
COPY templates templates/

RUN pwd
RUN ls -als 
RUN ls -als templates 

RUN go mod vendor
RUN go build -o updater .
 

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/updater .
COPY --from=builder /app/templates templates/
# # EXPOSE 8282
# ENTRYPOINT ["./updater"]
                             