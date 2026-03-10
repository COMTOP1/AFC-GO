FROM golang:1.25.5-alpine3.23 AS build

LABEL site="afc"
LABEL stage="builder"

WORKDIR /src/

ARG AFC_VERSION_ARG
ARG AFC_COMMIT_ARG

RUN apk --no-cache add ca-certificates

# Stores our dependencies
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy source
COPY . .

# Set build variables
RUN echo -n "-X 'main.Version=$AFC_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "' -X 'main.Commit=$AFC_COMMIT_ARG" >> ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

# Build the executable
RUN GOOS=linux GOARCH=amd64 go build -ldflags="$(cat ./ldflags)" -o /bin/afc

# Run the executable
FROM scratch
LABEL site="afc"
# Copy binary
COPY --from=build /bin/afc /bin/afc
ENTRYPOINT ["/bin/afc"]