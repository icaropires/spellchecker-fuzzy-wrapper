import subprocess
import json

def correct_string(tokens, vocab):
    # Spellcheck bin must be at same folder
    popen = subprocess.Popen(
        ("./spellchecker", tokens, vocab),
        stdout=subprocess.PIPE
    )
    popen.wait()

    output = popen.stdout.read().decode('utf-8').strip()

    corrections = json.loads(output)
    return ' '.join(map(corrections.get, tokens.split()))

tokens = "oii olal hello"
vocab = "oi ola hello"

print(correct_string(tokens, vocab))
