FROM golang:1.24

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o main ./cmd/main.go

#FROM debian:bullseye-slim
#WORKDIR /root/
#COPY --from=builder /app/main .
#COPY ./web ./web

EXPOSE 8081

CMD ["./main"]
