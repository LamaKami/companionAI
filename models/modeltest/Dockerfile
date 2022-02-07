FROM python:3.8-slim-buster

RUN pip install --upgrade pip \
 && pip install flask \
 && pip install install -U pip setuptools wheel \
 && pip install -U spacy==2.3.7 \
 && pip install -U spacy-lookups-data pyyaml \
 && python -m spacy download en

WORKDIR /app
COPY app.py ./app.py
COPY train.py ./train.py
COPY __init__.py ./__init__.py

CMD ["python3", "-m" , "flask", "run", "--host=0.0.0.0"]
