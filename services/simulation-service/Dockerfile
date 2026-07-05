FROM python:3.11-slim
WORKDIR /app
RUN apt-get update -qq && apt-get install -y -qq wget && rm -rf /var/lib/apt/lists/*
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY simulation_service ./simulation_service
COPY tests ./tests
EXPOSE 8094
ENV SIMULATION_SERVICE_HTTP_ADDR=:8094
CMD ["python", "-m", "uvicorn", "simulation_service.main:app", "--host", "0.0.0.0", "--port", "8094"]
