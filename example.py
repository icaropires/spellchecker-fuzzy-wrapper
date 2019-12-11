import subprocess
import json
import pickle
from typing import Iterable, Dict, Tuple, List


def get_corrections(tokens : str, vocab : str) -> Tuple[Dict[str,str], List[str]]:
    # Spellcheck bin must be at same folder
    popen = subprocess.Popen(
        ("./spellchecker", tokens, vocab),
        stdout=subprocess.PIPE
    )
    popen.wait()

    output = popen.stdout.read().decode('utf-8').strip()
    all_corrections = json.loads(output)

    corrections, unkowns = dict(), list()
    for token, correction in all_corrections.items():
        if correction:
            corrections[token] = correction
        else:
            unkowns.append(token)

    return corrections, unkowns


def correct_string(tokens : Iterable[str], corrections_dict : Dict[str, str]) -> str:
    return ' '.join(corrections_dict.get(t, '') for t in tokens)


def load_pickle(path : str) -> str:
    vocab = None
    with open(path, 'rb') as vocab_file:
        vocab_iterable = pickle.load(vocab_file)[:500]  # Problem with args size limit
        vocab = ' '.join(vocab_iterable)

    return vocab


tokens = load_pickle('tokens.pkl')
vocab = load_pickle('vocabulary.pkl')

corrections, unkowns = get_corrections((tokens), vocab)
correct_string(tokens.split(), corrections)

corrected = correct_string(tokens.split(), corrections)

print('Corrected output:', corrected)

if unkowns:
    print("Tokens excluded for not being recognized:",
          "\n".join(f'  - {u}' for u in unkowns),
           sep='\n')
