FROM python:3.11-slim
WORKDIR /app
RUN apt-get update -qq && apt-get install -y -qq wget && rm -rf /var/lib/apt/lists/*
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY oidc_edge ./oidc_edge
EXPOSE 8180
CMD ["python", "-m", "uvicorn", "oidc_edge.main:app", "--host", "0.0.0.0", "--port", "8180"]
