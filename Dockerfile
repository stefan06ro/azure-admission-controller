FROM alpine:3.13.2
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
