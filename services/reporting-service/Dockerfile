FROM python:3.11-slim
WORKDIR /app
RUN apt-get update -qq && apt-get install -y -qq wget && rm -rf /var/lib/apt/lists/*
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY reporting_service ./reporting_service
COPY migrations ./migrations
COPY tests ./tests
EXPOSE 8095
ENV REPORTING_SERVICE_HTTP_ADDR=:8095
CMD ["python", "-m", "uvicorn", "reporting_service.main:app", "--host", "0.0.0.0", "--port", "8095"]
