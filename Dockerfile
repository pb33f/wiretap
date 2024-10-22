FROM node:20-bookworm-slim AS uibuilder

ARG TARGETOS
ARG TARGETARCH

RUN echo "I am building ui for OS:$TARGETOS, ARCH:$TARGETARCH" > /log

# set up ui build
RUN mkdir -p /wt_build
WORKDIR /wt_build
COPY ui/ ./ui
RUN npm install yarn
WORKDIR /wt_build/ui

# run ui build
RUN yarn install
RUN yarn build

FROM golang:1.23-bookworm AS gobuilder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /work

# copy ui build from previous stage
COPY --from=uibuilder /wt_build/ui/dist ./ui/dist

COPY . ./

RUN echo "I am building go for GOOS:$TARGETOS, GOARCH:$TARGETARCH" > /log

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -v -o /wiretap wiretap.go

FROM golang:1.23-alpine AS runner

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "I am running on $TARGETPLATFORM, was built on $BUILDPLATFORM" > /log

# copy only the built binary
COPY --from=gobuilder /wiretap /wiretap

ENV PATH=$PATH:/work/bin

ENTRYPOINT ["wiretap"]