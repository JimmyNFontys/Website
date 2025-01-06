# Begin met de officiële Go base image.
FROM golang:1.22.3

# Stel de werkdirectory in de container in.
WORKDIR /webapp

# Kopieer de go.mod en go.sum om de afhankelijkheden te beheren.
COPY go.mod .
COPY go.sum .

# Download alle afhankelijkheden.
RUN go mod download

# Kopieer de broncode van de applicatie naar de container.
COPY main.go .
COPY config.txt .
COPY handlers/ handlers/
COPY locales/ locales/
COPY scripts/ scripts/
COPY templates/ templates/
COPY image/ image/

# Bouw de Go app als een binary.
RUN go build -o Webapp1 .

# Deze haven wordt door de applicatie binnen de container gebruikt.
EXPOSE 8080

# Run de gecompileerde binary.
CMD ["./Webapp1"]