# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Contact maintainer with any issues you encounter
MAINTAINER Richard Knop <risoknop@gmail.com>

# Cd into the api code directory
WORKDIR /go/src/github.com/RichardKnop/example-api

# Create a new unprivileged user
RUN useradd --user-group --shell /bin/false www

# Chown /go/src/github.com/RichardKnop/example-api/ to www user
RUN chown -R www:www /go/src/github.com/RichardKnop/example-api/

# Use the unprivileged user
USER www

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/RichardKnop/example-api

# Install the api program
RUN go install github.com/RichardKnop/example-api

# Set environment variables
ENV PATH /go/bin:$PATH

# Copy the docker-entrypoint.sh script and use it as entrypoint
COPY ./docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

# Document that the service listens on port 8080.
EXPOSE 8080
