FROM postgres:16-alpine

# Set environment variables
ENV POSTGRES_DB=socialforge
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres

# Copy custom initialization scripts if needed
COPY ./init.sql /docker-entrypoint-initdb.d/

# Expose the PostgreSQL port
EXPOSE 5432

# The default command will start PostgreSQL
CMD ["postgres"] 