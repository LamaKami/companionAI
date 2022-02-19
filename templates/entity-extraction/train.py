from __future__ import unicode_literals, print_function
import random
from pathlib import Path
import spacy
from tqdm import tqdm
import json
import yaml


def load_config(path):
    return yaml.safe_load(open(path))


def load_data(path):
    f = open(path)
    trainings_data = json.load(f)

    data = []

    for element in trainings_data["dataPoints"]:
        data.append((element["sentence"],{'entities': [(x["start"], x["end"], x["label"]) for x in element["entities"]]}))
    return data


def check_trainingsData():
    config = load_config("/mnt/data/config.yml")
    return load_data(config["trainingsData"])


def train_model(train_data):
    config = load_config("/mnt/data/config.yml")

    n_iter = config["n_iter"]

    nlp = spacy.blank(config["modelLanguage"])
    print(f"Created blank {config['modelLanguage']} model")

    ner = nlp.create_pipe('ner')
    nlp.add_pipe(ner, last=True)

    # train
    for _, annotations in train_data:
        for ent in annotations.get('entities'):
            ner.add_label(ent[2])

    other_pipes = [pipe for pipe in nlp.pipe_names if pipe != 'ner']
    with nlp.disable_pipes(*other_pipes):  # only train NER
        optimizer = nlp.begin_training()
        for itn in range(n_iter):
            random.shuffle(train_data)
            losses = {}
            for text, annotations in tqdm(train_data):
                nlp.update(
                    [text],
                    [annotations],
                    drop=config["drop"],
                    sgd=optimizer,
                    losses=losses)
            yield "data: " + "iterations: " + str(itn+1) + "/" + str(n_iter) + " " + str(losses) + "\n"

    nlp.meta['name'] = 'companionAI-ner'
    nlp.to_disk(f"/mnt/model-{config['currentVersion']}")