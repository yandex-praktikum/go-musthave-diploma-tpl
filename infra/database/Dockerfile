# Рекомендуется 17 версия, так как 18 еще не вышла

FROM postgres:alpine

EXPOSE 5550

COPY HEALTHCHECK.sh /HEALTHCHECK.sh
RUN chmod +x /HEALTHCHECK.sh

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/bin/sh", "/HEALTHCHECK.sh"]

CMD ["postgres"]