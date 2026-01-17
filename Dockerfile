FROM golang:1.24.4

ENV TZ=Asia/Jakarta

WORKDIR /app

# Copy dependency definitions
COPY go.mod go.sum ./
RUN go mod download

# Copy app source, .env files, and .git-branch file
COPY . .

# (Optional) log copied files for debugging
RUN ls -la .env*

# Build binary
RUN go build -o main .

# Port from env file will be read by the Go app
EXPOSE 47001

# Run the app
CMD ["./main"]
