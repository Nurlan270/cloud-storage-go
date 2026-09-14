FROM minio/minio:RELEASE.2025-09-07T16-13-09Z-cpuv1

COPY ./docker/minio/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]