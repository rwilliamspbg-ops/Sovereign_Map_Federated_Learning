#!/usr/bin/env python
import re
import os

fpath = os.path.join(os.path.dirname(__file__), 'bft_week1_realistic.py')

with open(fpath, 'r', encoding='utf-8') as f:
    content = f.read()

# Replace all unicode checkmarks
replacements = {
    '✓': '[OK]',
    '✗': '[XX]',
    '⚠️': '[WARN]',
    '⚠': '[WARN]',
}

for old, new in replacements.items():
    content = content.replace(old, new)

with open(fpath, 'w', encoding='utf-8') as f:
    f.write(content)

print('Fixed unicode characters')
