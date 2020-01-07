import subprocess
import json
import pickle
from typing import Iterable, Dict, Tuple, List


TOKENS_FILE = "tokens.txt"
VOCAB_FILE = "vocabulary.txt"
CORRECTED_TOKENS_FILE = "corrections.pkl"


def get_corrections(tokens_file: str, vocab_file: str) -> Tuple[Dict[str, str], List[str]]:
    depth = 2  # More, slower, more false positives, more detailed analysis

    # Spellcheck bin must be at same folder
    popen = subprocess.Popen(
        ("./spellchecker", str(depth), tokens_file, vocab_file),
        stdout=subprocess.PIPE
    )

    dump = ''
    last_processed = ''
    loading, li = ('/', '-', '\\', '|'), 0
    print('Generating corrections:')
    while True:
        line = popen.stdout.readline().decode('utf-8').strip()

        if not line and popen.poll() is not None:
            break

        if line:
            dump += line
            last_processed = line[1].upper()

            if ord(last_processed) >= 65 or ord(last_processed) <= 90:
                li = (li + 1) % len(loading)
                print(f'  Processing words with: {last_processed} {loading[li]}',
                      end='\r')
    print()

    all_corrections = json.loads(dump)

    corrections, unkowns = dict(), list()
    for token, correction in all_corrections.items():
        if correction:
            corrections[token] = correction
        else:
            unkowns.append(token)

    return corrections, unkowns


def correct_string(tokens: Iterable[str], corrections_dict: Dict[str, str]) -> str:
    return ' '.join(corrections_dict.get(t, '') for t in tokens)


corrections, unkowns = get_corrections(TOKENS_FILE, VOCAB_FILE)

tokens = None
with open(TOKENS_FILE, 'r') as tf:
    tokens = tf.readlines()

corrected = correct_string(tokens, corrections)

with open(CORRECTED_TOKENS_FILE, 'wb') as corrected_file:
    pickle.dump(corrections, corrected_file)

print(f'Corrected string: {corrected}')
print(f'Corrected tokens saved to {CORRECTED_TOKENS_FILE}:', corrected)

unkowns_str = "\n".join(f'  - {u}' for u in unkowns if u)
if unkowns:
    print("Tokens excluded for not being recognized:",
          unkowns_str,
          sep='\n')

with open('unknowns.txt', 'w') as f:
    f.write(unkowns_str)
