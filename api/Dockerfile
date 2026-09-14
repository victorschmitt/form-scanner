FROM ghcr.io/hybridgroup/opencv:4.13.0

WORKDIR /src
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=1
ENV PORT=8080

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/form-scanner .

EXPOSE 8080
CMD ["form-scanner"]
