# Etapa base
FROM golang:1.22-alpine

# Instalar Air y dependencias
RUN apk add --no-cache git curl build-base \
    && go install github.com/air-verse/air@latest

# Crear carpeta de trabajo
WORKDIR /app

# Copiar los archivos de Go y descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Puerto donde correrá Gin
EXPOSE 8080

# Comando por defecto
CMD ["air"]
