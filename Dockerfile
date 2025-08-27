FROM golang

WORKDIR /app

COPY . .

RUN go build -o my_app

EXPOSE 8081

CMD ["./my_app"]