FROM python:3.12.5-alpine3.19

WORKDIR /home/andrew

ENV SERVE_PORT=8000

RUN pip install boto3

COPY recipe-maker.py .

CMD ["python3", "recipe-maker.py"]
