FROM alpine:3.13.5
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
