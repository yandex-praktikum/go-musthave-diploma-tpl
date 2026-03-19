# Используем абсолютно пустой образ
FROM scratch

# Копируем наш предварительно собранный статичный бинарник
COPY ./cmd/gophermart/gophermart /gophermart
COPY ./.env /.env

# Если сервис использует SSL (https), могут понадобиться сертификаты
# Их можно скопировать из хостовой системы, если нужно:
# COPY /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем конфиг .env (если он используется для параметров)
COPY .env /.env

# Экспонируем порт
EXPOSE 8080

# Запуск
ENTRYPOINT ["/gophermart"]
