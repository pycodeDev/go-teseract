FROM golang:1.23

# Install tesseract and dependencies
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    libtesseract-dev \
    libleptonica-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set working dir
WORKDIR /app

# Copy Go files and build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ocrservice main.go

EXPOSE 8003

CMD ["./ocrservice"]
