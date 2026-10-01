FROM node:24-alpine AS client

WORKDIR /src/
RUN corepack enable

COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

COPY tsconfig.json tsconfig.app.json tsconfig.node.json vite.config.ts eslint.config.js .prettierrc ./
COPY scripts ./scripts
COPY client ./client
# CI lints; the image build only compiles.
RUN BUILD_CLIENT_SKIP_LINT=true yarn build:client


FROM golang:1.26.7-alpine3.24 AS build

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

# Embed the web client built in the first stage
COPY --from=client /src/build/client/ ./server/cmd/afc/ui/

# Set build variables
RUN echo -n "-X 'main.Version=$AFC_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "' -X 'main.Commit=$AFC_COMMIT_ARG" >> ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

# Build the executable
RUN GOOS=linux GOARCH=amd64 GOEXPERIMENT=runtimesecret go build -ldflags="$(cat ./ldflags)" -o /bin/afc ./server/cmd/afc
RUN GOOS=linux GOARCH=amd64 go build -o /bin/migrates3 ./server/cmd/migrates3

# Run the executable
FROM alpine:3.24 AS run

RUN apk add --no-cache ca-certificates

LABEL site="afc"
# Copy binary
COPY --from=build /bin/afc /bin/afc
COPY --from=build /bin/migrates3 /bin/migrates3
ENTRYPOINT ["/bin/afc"]