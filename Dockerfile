FROM golang:1.12-stretch

# Install npm and parcel
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash - && \
    apt-get install -y nodejs && \
    apt-get clean && \
    npm install -g parcel-bundler
    
# Copy project into docker instance
COPY . /app
WORKDIR /app

# Get the go app
RUN go get -u github.com/jbowens/codenames

# Build backend and frontend 
RUN go build cmd/codenames/main.go && \
    cd /app/frontend/ && \
    npm install && \
    sh build.sh

# Expose 9091 port
EXPOSE 9091/tcp

# Set entrypoint command
CMD /app/main
