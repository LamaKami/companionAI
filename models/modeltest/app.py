import spacy
import flask
from flask import Flask, request, Response
from train import train_model, check_trainingsData

NLP = None

app = Flask(__name__)


@app.route('/predict', methods=['POST'])
def predict():
    content = request.json
    try:
        sentence = NLP(content['sentence'])
    except:
        return 'Could not predict the data. Did you already load the model?', 400
    elements = {}
    for ent in sentence.ents:
        elements[ent.label_] = ent.text
    return elements


@app.route('/load/<version>', methods=['GET'])
def load(version):
    global NLP
    output_dir = f"/mnt/model-{version}"
    try:
        NLP = spacy.load(output_dir)
        return flask.Response(status=201)
    except:
        return 'Model could not be loaded. Did you already train?', 400


@app.route('/train', methods=['GET'])
def train():
    try:
        data = check_trainingsData()
        return Response(train_model(data), mimetype='text/event-stream')
    except:
        return 'Did you create some trainings data?', 400
