FROM golang:1.23.7

WORKDIR /opt/app

# Installa Air per il live reload
RUN go install github.com/air-verse/air@latest

COPY . .

# Usa il file .env per passare variabili d’ambiente
ARG BE_HOST=""
ARG BE_PORT="9090"
ARG DB_HOST=""
ARG DB_USER=""
ARG DB_PASS=""
ARG DB_NAME=""
ARG DB_PORT="5432"

ENV BE_HOST=${BE_HOST}
ENV BE_PORT=${BE_PORT}
ENV DB_HOST=${DB_HOST}
ENV DB_USER=${DB_USER}
ENV DB_PASS=${DB_PASS}
ENV DB_NAME=${DB_NAME}
ENV DB_PORT=${DB_PORT}

EXPOSE 9090 4000

# Avvia l'app con Air per il live reload
# CMD ["air -c /opt/app/air.toml"]
