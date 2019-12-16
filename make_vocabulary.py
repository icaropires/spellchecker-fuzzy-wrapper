import re
import pickle
from sys import argv, exit

import numpy as np

from pre_processing.pre_processing import CorpusHandler

def get_words(text):
    return re.findall(r'[A-ZÁÉÍÓÚÂÊÔÀÁÉÍÓÚÂÊÔÀ]?[a-záéíóúâêôàáéíóúâãõêôàç\-]*[a-záéíóúâêôàáéíóúâãõêôà][.!?,:;"\'()]?\s', text)

def clean(w):
    endings = ('.', ',', '!', '?', ':', ')', '(', '"', "'", ';')

    w = w.lower().strip()
    for e in endings:
        w = w.strip(e)

    return w

def apply_preprocessing(text):
    import nltk
    from nltk.corpus import stopwords
    import unidecode
    nltk.download("stopwords")
    nltk.download('rslp')
    stopwords = stopwords.words("portuguese")
    stopwords = np.array(stopwords, dtype="unicode")
    stopwords = [unidecode.unidecode(x) for x in stopwords]

    text = CorpusHandler.clean_email(text)
    text = CorpusHandler.clean_site(text)
    text = CorpusHandler.transform_token(text)
    text = CorpusHandler.clean_special_chars(text)
    text = CorpusHandler.remove_letter_number(text)
    text = CorpusHandler.clean_number(text)
    text = CorpusHandler.clean_alphachars(text)
    text = CorpusHandler.clean_document(text)
    text = CorpusHandler.remove_stop_words(text, stopwords)
    text = text.split()
    text = ' '.join(x for x in text if len(x) > 2)
    text = CorpusHandler.clean_spaces(text)

    return text

if len(argv) < 2:
    print('Incorrect usage! Usage: python3 spellchecker.py [corpus]')
    exit(1)

content = open(argv[1]).read()
words = get_words(content)

text = ' '.join(set(map(clean, words)))
words = apply_preprocessing(text).split()

vocabulary_path = 'vocabulary.txt'
with open(vocabulary_path, 'w') as vocabulary_file:
    vocabulary_file.write('\n'.join(words))
print(f'Vocabulary saved to {vocabulary_path}!')
