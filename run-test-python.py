#!/usr/bin/env python3
import os
import runpy
import sys

script = os.path.join(
    os.path.dirname(__file__), "tests", "scripts", "python", "run-test-python.py"
)
sys.argv[0] = script
runpy.run_path(script, run_name="__main__")
