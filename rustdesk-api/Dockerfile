FROM alpine

ARG BUILDARCH
WORKDIR /app
RUN apk add --no-cache tzdata sqlite
COPY ./${BUILDARCH}/release /app/
VOLUME /app/data

EXPOSE 21114
CMD ["./apimain"]
