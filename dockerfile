FROM golang

WORKDIR /app

COPY go.mod .
COPY main.go .

RUN go build -o main .

ENTRYPOINT [ "/app/main" ]