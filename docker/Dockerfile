FROM redis:8.2.3

# Install OpenMP runtime required by FalkorDB
RUN apt-get update \
    && apt-get install -y --no-install-recommends libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# Create a directory for Redis modules
RUN mkdir -p /usr/lib/redis/modules

# Copy the module from the deps directory (relative to build context)
COPY deps/falkordb-x64.so /usr/lib/redis/modules/falkordb-x64.so

# Redis requires execute permission on modules
RUN chmod 755 /usr/lib/redis/modules/falkordb-x64.so

# Start Redis with the module loaded
CMD ["redis-server", "--loadmodule", "/usr/lib/redis/modules/falkordb-x64.so"]
