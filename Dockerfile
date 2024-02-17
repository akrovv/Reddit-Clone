FROM golang:alpine AS builder

RUN mkdir /reddit
WORKDIR /reddit

COPY . .
COPY .env .

RUN go mod download && go build cmd/redditclone/main.go

FROM alpine

WORKDIR /api

RUN mkdir front

COPY --from=builder /reddit/front ./front
COPY --from=builder /reddit/basic_model.conf .
COPY --from=builder /reddit/basic_policy.csv .
COPY --from=builder /reddit/.env .
COPY --from=builder /reddit/main .

EXPOSE 8080

CMD [ "./main" ]