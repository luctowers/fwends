FROM redis:alpine
# TODO: use a proper acl or config file
CMD ["sh", "-c", "exec redis-server --requirepass \"$REDIS_PASSWORD\""]
