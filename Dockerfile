	# Use the official Golang image as the base image
	FROM ubuntu:latest
	 
	# Set the working directory inside the container
	WORKDIR /app
	 
	# Copy the go.mod and go.sum files to the working directory
	##COPY go.mod go.sum ./
	COPY templates ./templates
	COPY assets ./assets
	COPY config ./config
	COPY manifest.json  ./
	 
	# Download and install the Go dependencies
	##RUN go mod download
	 
	# Copy the rest of the application source code to the working directory
	COPY /tmp/arcwiki ./
	# Build the Go application
	##RUN go build -o arcwiki

	# Set environment variables for configuration
	ENV PORT=8080
	ENV DB_HOST=localhost
	ENV DB_PORT=5432
	ENV LOG_LEVEL=info
	 
	# Set a label for the maintainer
	LABEL maintainer="Edward Stock <edd@eddland.co.uk>"
	 
	
	 
	# Expose the port on which the application will run
	EXPOSE $PORT
	 
	# Run the Go application
	CMD ["./arcwiki"]
