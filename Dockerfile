	# Use the official Golang image as the base image
	FROM golang:latest
	 
	# Set the working directory inside the container
	WORKDIR /app
	 
	# Copy the go.mod and go.sum files to the working directory
	COPY go.mod go.sum ./
	 
	# Download and install the Go dependencies
	RUN go mod download
	 
	# Copy the rest of the application source code to the working directory
	COPY . .
	 
	# Set environment variables for configuration
	ENV PORT=8080
	ENV DB_HOST=localhost
	ENV DB_PORT=5432
	 
	# Set a default value for the environment variable
	ENV LOG_LEVEL=info
	 
	# Set a label for the maintainer
	LABEL maintainer="Edward Stock <edd@eddland.co.uk>"
	 
	# Build the Go application
	RUN go build -o arcwiki
	 
	# Expose the port on which the application will run
	EXPOSE $PORT
	 
	# Run the Go application
	CMD ["./arcwiki"]
