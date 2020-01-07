#!/bin/env python3

from sys import argv, exit
from typing import List

from pre_processing import CorpusHandler


def get_words(text: str) -> List[str]:
    text = text.lower()
    text = CorpusHandler.remove_small_big_words(text)

    return CorpusHandler.tokenize(text)


if __name__ == '__main__':
    if len(argv) < 2:
        print('Incorrect usage! Usage: python3 make_vocabulary.py [corpus]')
        exit(1)

    with open(argv[1]) as corpus_file:
        content = corpus_file.read()

    words = set(get_words(content))

    vocabulary_path = 'vocabulary.txt'
    with open(vocabulary_path, 'w') as vocabulary_file:
        vocabulary_file.write('\n'.join(words))

    print(f'Vocabulary saved to {vocabulary_path}!')
