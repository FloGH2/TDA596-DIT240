FROM golang:1.25-alpine

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod ./

COPY server/ ./server/
COPY proxy/ ./proxy/

RUN cd server && go build -o /app/http_server .
RUN cd proxy && go build -o /app/proxy .

RUN mkdir -p uploads

EXPOSE 8080 9090

CMD ["echo", "Use docker-compose to start individual services"]
