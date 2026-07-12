FROM python:3.11-slim@sha256:e031123e3d85762b141ad1cbc56452ba69c6e722ebf2f042cc0dc86c47c0d8b3
WORKDIR /app
RUN apt-get update -qq && apt-get install -y -qq wget && rm -rf /var/lib/apt/lists/*
COPY requirements.txt .
RUN pip install --no-cache-dir --require-hashes -r requirements.txt
COPY simulation_service ./simulation_service
COPY tests ./tests
EXPOSE 8094
ENV SIMULATION_SERVICE_HTTP_ADDR=:8094
CMD ["python", "-m", "uvicorn", "simulation_service.main:app", "--host", "0.0.0.0", "--port", "8094"]
